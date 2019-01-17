package service

import (
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) updateSensorCfg(config interface{}) {
	cfg, _ := ds.ToSensorConf(config)

	oldSensor, _ := database.GetSensorConfig(s.db, cfg.Mac)
	if oldSensor == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	database.UpdateSensorConfig(s.db, *cfg)
	//Get correspnding switchMac
	sensor, _ := database.GetSensorConfig(s.db, cfg.Mac)
	if sensor == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	if sensor.Group != nil {
		if oldSensor.Group != sensor.Group {
			if oldSensor.Group != nil {
				rlog.Info("Update old group", *oldSensor.Group)
				gr, _ := database.GetGroupConfig(s.db, *oldSensor.Group)
				if gr != nil {
					for i, v := range gr.Sensors {
						if v == sensor.Mac {
							gr.Sensors = append(gr.Sensors[:i], gr.Sensors[i+1:]...)
							break
						}
					}
					rlog.Info("Old group will be ", gr.Sensors)
					s.updateGroupCfg(gr)
				}
			}
			rlog.Info("Update new group", *sensor.Group)
			grNew, _ := database.GetGroupConfig(s.db, *sensor.Group)
			if grNew != nil {
				grNew.Sensors = append(grNew.Sensors, cfg.Mac)
				rlog.Info("new group will be", grNew.Sensors)
				s.updateGroupCfg(grNew)
			}
		}
	}

	url := "/write/switch/" + sensor.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = sensor.SwitchMac
	switchSetup.SensorsConfig = make(map[string]ds.SensorConf)

	switchSetup.SensorsConfig[cfg.Mac] = *cfg

	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + sensor.SwitchMac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + sensor.SwitchMac + " on topic: " + url + " dump:" + dump)
	}
}