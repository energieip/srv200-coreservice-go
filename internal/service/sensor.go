package service

import (
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) updateGroupSensor(oldSensor ds.SensorSetup, sensor ds.SensorSetup) {
	if sensor.Group != nil {
		if oldSensor.Group != sensor.Group {
			if oldSensor.Group != nil {
				rlog.Info("Update old group", *oldSensor.Group)
				gr, _ := database.GetGroupConfig(s.db, *oldSensor.Group)
				if gr != nil {
					sensors := []string{}
					for _, v := range gr.Sensors {
						if v != sensor.Mac {
							sensors = append(sensors, v)
						}
					}
					gr.Sensors = sensors
					rlog.Info("Old group will be ", gr.Sensors)
					s.updateGroupCfg(gr)
				}
			}
			rlog.Info("Update new group", *sensor.Group)
			grNew, _ := database.GetGroupConfig(s.db, *sensor.Group)
			if grNew != nil {
				grNew.Sensors = append(grNew.Sensors, sensor.Mac)
				rlog.Info("new group will be", grNew.Sensors)
				s.updateGroupCfg(grNew)
			}
		}
	}
}

func (s *CoreService) sendSwitchSensorSetup(elt ds.SensorSetup) {
	if elt.SwitchMac == "" {
		return
	}

	url := "/write/switch/" + elt.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = elt.SwitchMac
	switchSetup.SensorsSetup = make(map[string]ds.SensorSetup)
	switchSetup.SensorsSetup[elt.Mac] = elt

	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

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
	s.updateGroupSensor(*oldSensor, *sensor)

	url := "/write/switch/" + sensor.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = sensor.SwitchMac
	switchSetup.SensorsConfig = make(map[string]ds.SensorConf)
	switchSetup.SensorsConfig[cfg.Mac] = *cfg

	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) updateSensorSetup(config interface{}) {
	byLbl := false
	cfg, _ := ds.ToSensorSetup(config)
	if cfg == nil {
		return
	}

	oldSensor, _ := database.GetSensorConfig(s.db, cfg.Mac)
	if oldSensor == nil && cfg.Label != nil {
		oldSensor, _ = database.GetSensorLabelConfig(s.db, *cfg.Label)
		if oldSensor != nil {
			//it means that the IFC has been uploaded but the MAC is unknown
			byLbl = true
		}
	}

	if oldSensor != nil {
		s.updateGroupSensor(*oldSensor, *cfg)
	}
	if byLbl {
		database.UpdateSensorLabelSetup(s.db, *cfg)
	} else {
		database.UpdateSensorSetup(s.db, *cfg)
	}
	//Get correspnding switchMac
	sensor, _ := database.GetSensorConfig(s.db, cfg.Mac)
	if sensor == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}
	s.sendSwitchSensorSetup(*cfg)
}

func (s *CoreService) updateSensorLabelSetup(config interface{}) {
	cfg, _ := ds.ToSensorSetup(config)
	if cfg == nil || cfg.Label == nil {
		return
	}

	oldSensor, _ := database.GetSensorLabelConfig(s.db, *cfg.Label)
	if oldSensor != nil {
		s.updateGroupSensor(*oldSensor, *cfg)
	}

	database.UpdateSensorLabelSetup(s.db, *cfg)
	//Get correspnding switchMac
	sensor, _ := database.GetSensorConfig(s.db, cfg.Mac)
	if sensor == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}
	s.sendSwitchSensorSetup(*cfg)
}
