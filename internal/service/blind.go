package service

import (
	"github.com/energieip/common-components-go/pkg/dblind"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) updateBlindCfg(config interface{}) {
	cfg, _ := dblind.ToBlindConf(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldBlind, _ := database.GetBlindConfig(s.db, cfg.Mac)
	if oldBlind == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	database.UpdateBlindConfig(s.db, *cfg)
	//Get corresponding switchMac
	blind, _ := database.GetBlindConfig(s.db, cfg.Mac)
	if blind == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	if blind.Group != nil {
		if oldBlind.Group != blind.Group {
			if oldBlind.Group != nil {
				rlog.Info("Update old group", *oldBlind.Group)
				gr, _ := database.GetGroupConfig(s.db, *oldBlind.Group)
				if gr != nil {
					for i, v := range gr.Blinds {
						if v == blind.Mac {
							gr.Blinds = append(gr.Blinds[:i], gr.Blinds[i+1:]...)
							break
						}
					}
					rlog.Info("Old group will be ", gr.Blinds)
					s.updateGroupCfg(gr)
				}
			}
			rlog.Info("Update new group", *blind.Group)
			grNew, _ := database.GetGroupConfig(s.db, *blind.Group)
			if grNew != nil {
				grNew.Blinds = append(grNew.Blinds, cfg.Mac)
				rlog.Info("new group will be", grNew.Blinds)
				s.updateGroupCfg(grNew)
			}
		}
	}
	url := "/write/switch/" + blind.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = blind.SwitchMac
	switchSetup.BlindsConfig = make(map[string]dblind.BlindConf)

	switchSetup.BlindsConfig[cfg.Mac] = *cfg

	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + blind.SwitchMac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + blind.SwitchMac + " on topic: " + url + " dump:" + dump)
	}
}

func (s *CoreService) sendBlindCmd(cmdBlind interface{}) {
	cmd, _ := core.ToBlindCmd(cmdBlind)
	if cmd == nil {
		rlog.Error("Cannot parse cmd")
		return
	}
	//Get correspnding switchMac
	driver, _ := database.GetBlindConfig(s.db, cmd.Mac)
	if driver == nil {
		rlog.Error("Cannot find config for " + cmd.Mac)
		return
	}
	url := "/write/switch/" + driver.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = driver.SwitchMac
	switchSetup.BlindsConfig = make(map[string]dblind.BlindConf)

	cfg := dblind.BlindConf{
		Mac:    cmd.Mac,
		Blind1: &cmd.Blind1,
		Blind2: &cmd.Blind2,
		Slat1:  &cmd.Slat1,
		Slat2:  &cmd.Slat2,
	}
	rlog.Info("Ready to send ", cfg)
	rlog.Info("To switch", driver.SwitchMac)
	switchSetup.BlindsConfig[cmd.Mac] = cfg

	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + driver.SwitchMac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + driver.SwitchMac + " on topic: " + url + " dump:" + dump)
	}
}