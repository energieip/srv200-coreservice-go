package service

import (
	"strings"

	gm "github.com/energieip/common-components-go/pkg/dgroup"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) replaceDriver(driver interface{}) {
	replace, _ := core.ToReplaceDriver(driver)
	if replace == nil {
		rlog.Error("Cannot parse replace driver")
		return
	}

	project := database.GetProjectByFullMac(s.db, replace.OldFullMac)
	if project == nil {
		rlog.Error("Unkown old driver")
		return
	}

	project.FullMac = &replace.NewFullMac
	oldMac := ""
	if project.Mac != nil {
		oldMac = *project.Mac
	}

	submac := strings.SplitN(replace.NewFullMac, ":", 4)
	mac := submac[len(submac)-1]
	project.Mac = &mac

	//update driver tables
	if project.ModelName != nil {
		refModel := *project.ModelName
		if strings.HasPrefix(refModel, "led") {
			oldDriver, _ := database.GetLedConfig(s.db, oldMac)
			if oldDriver == nil {
				rlog.Error("Cannot find Led " + oldMac + " in database")
				return
			}

			err := database.SwitchLedConfig(s.db, oldMac, replace.OldFullMac, *project.Mac, *project.FullMac)
			if err != nil {
				rlog.Error("Cannot update Led database", err)
				return
			}

			//send remove reset old driver configuration to the switch
			switchConf := sd.SwitchConfig{}
			switchConf.Mac = oldDriver.SwitchMac
			switchConf.LedsSetup[oldDriver.Mac] = *oldDriver
			s.sendSwitchRemoveConfig(switchConf)

			// update group configuration
			// send update to all switch where this group is running
			groupCfg, _ := database.GetGroupConfig(s.db, *oldDriver.Group)
			newLeds := []string{}
			for _, led := range groupCfg.Leds {
				if led != oldMac {
					newLeds = append(newLeds, led)
				}
			}
			groupCfg.Leds = newLeds

			database.UpdateGroupConfig(s.db, *groupCfg)
			newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
			for sw := range newSwitch {
				url := "/write/switch/" + sw + "/update/settings"
				switchSetup := sd.SwitchConfig{}
				switchSetup.Mac = sw
				switchSetup.Groups = make(map[int]gm.GroupConfig)
				switchSetup.Groups[groupCfg.Group] = *groupCfg
				dump, _ := switchSetup.ToJSON()
				s.server.SendCommand(url, dump)
			}

		} else {
			if strings.HasPrefix(refModel, "bld") {
				oldDriver, _ := database.GetBlindConfig(s.db, oldMac)
				if oldDriver == nil {
					rlog.Error("Cannot find Blind " + oldMac + " in database")
					return
				}

				err := database.SwitchBlindConfig(s.db, oldMac, replace.OldFullMac, *project.Mac, *project.FullMac)
				if err != nil {
					rlog.Error("Cannot update Blind database", err)
					return
				}

				//send remove reset old driver configuration to the switch
				switchConf := sd.SwitchConfig{}
				switchConf.Mac = oldDriver.SwitchMac
				switchConf.BlindsSetup[oldDriver.Mac] = *oldDriver
				s.sendSwitchRemoveConfig(switchConf)

				// update group configuration
				// send update to all switch where this group is running
				groupCfg, _ := database.GetGroupConfig(s.db, *oldDriver.Group)
				newBlinds := []string{}
				for _, blind := range groupCfg.Blinds {
					if blind != oldMac {
						newBlinds = append(newBlinds, blind)
					}
				}
				groupCfg.Blinds = newBlinds

				database.UpdateGroupConfig(s.db, *groupCfg)
				newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
				for sw := range newSwitch {
					url := "/write/switch/" + sw + "/update/settings"
					switchSetup := sd.SwitchConfig{}
					switchSetup.Mac = sw
					switchSetup.Groups = make(map[int]gm.GroupConfig)
					switchSetup.Groups[groupCfg.Group] = *groupCfg
					dump, _ := switchSetup.ToJSON()
					s.server.SendCommand(url, dump)
				}
			} else {
				//sensor
				oldDriver, _ := database.GetSensorConfig(s.db, oldMac)
				if oldDriver == nil {
					rlog.Error("Cannot find Blind " + oldMac + " in database")
					return
				}

				err := database.SwitchSensorConfig(s.db, oldMac, replace.OldFullMac, *project.Mac, *project.FullMac)
				if err != nil {
					rlog.Error("Cannot update Blind database", err)
					return
				}

				//send remove reset old driver configuration to the switch
				switchConf := sd.SwitchConfig{}
				switchConf.Mac = oldDriver.SwitchMac
				switchConf.SensorsSetup[oldDriver.Mac] = *oldDriver
				s.sendSwitchRemoveConfig(switchConf)

				// update group configuration
				// send update to all switch where this group is running
				groupCfg, _ := database.GetGroupConfig(s.db, *oldDriver.Group)
				newSensors := []string{}
				for _, sensor := range groupCfg.Sensors {
					if sensor != oldMac {
						newSensors = append(newSensors, sensor)
					}
				}
				groupCfg.Sensors = newSensors

				database.UpdateGroupConfig(s.db, *groupCfg)
				newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
				for sw := range newSwitch {
					url := "/write/switch/" + sw + "/update/settings"
					switchSetup := sd.SwitchConfig{}
					switchSetup.Mac = sw
					switchSetup.Groups = make(map[int]gm.GroupConfig)
					switchSetup.Groups[groupCfg.Group] = *groupCfg
					dump, _ := switchSetup.ToJSON()
					s.server.SendCommand(url, dump)
				}
			}
		}
	}

	//update project
	err := database.SaveProject(s.db, *project)
	if err != nil {
		rlog.Error("Cannot saved new project configuration")
		return
	}
}