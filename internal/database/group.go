package database

import (
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/pconst"
)

//SaveGroupConfig dump group config in database
func SaveGroupConfig(db Database, cfg gm.GroupConfig) error {
	criteria := make(map[string]interface{})
	criteria["Group"] = cfg.Group
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbGroups, criteria)
}

//RemoveGroupConfig remove group config in database
func RemoveGroupConfig(db Database, grID int) error {
	criteria := make(map[string]interface{})
	criteria["Group"] = grID
	return db.DeleteRecord(pconst.DbConfig, pconst.TbGroups, criteria)
}

//GetGroupConfig return the group configuration
func GetGroupConfig(db Database, grID int) (*gm.GroupConfig, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Group"] = grID
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbGroups, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	gr, err := gm.ToGroupConfig(stored)
	if err != nil {
		return nil, dbID
	}
	return gr, dbID
}

//GetGroupSwitchs return the corresponding running switch list
func GetGroupSwitchs(db Database, grID int) map[string]bool {
	switchs := make(map[string]bool)
	criteria := make(map[string]interface{})
	criteria["Group"] = grID
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbGroups, criteria)
	if err != nil || stored == nil {
		return nil
	}
	gr, err := gm.ToGroupConfig(stored)
	if err != nil {
		return nil
	}
	for _, ledMac := range gr.Leds {
		led, _ := GetLedConfig(db, ledMac)
		if led == nil {
			continue
		}
		switchs[led.SwitchMac] = true
	}
	for _, blindMac := range gr.Blinds {
		blind, _ := GetBlindConfig(db, blindMac)
		if blind == nil {
			continue
		}
		switchs[blind.SwitchMac] = true
	}
	for _, hvacMac := range gr.Hvacs {
		hvac, _ := GetHvacConfig(db, hvacMac)
		if hvac == nil {
			continue
		}
		switchs[hvac.SwitchMac] = true
	}
	return switchs
}

//UpdateGroupConfig update group config in database
func UpdateGroupConfig(db Database, config gm.GroupConfig) error {
	setup, dbID := GetGroupConfig(db, config.Group)
	if setup == nil || dbID == "" {
		return NewError("Group " + string(config.Group) + " not found")
	}

	if config.Leds != nil {
		setup.Leds = config.Leds
	}

	if config.Sensors != nil {
		setup.Sensors = config.Sensors
	}

	if config.Blinds != nil {
		setup.Blinds = config.Blinds
	}

	if config.Hvacs != nil {
		setup.Hvacs = config.Hvacs
	}

	if config.Nanosenses != nil {
		setup.Nanosenses = config.Nanosenses
	}

	if config.FriendlyName != nil {
		setup.FriendlyName = config.FriendlyName
	}

	if config.CorrectionInterval != nil {
		setup.CorrectionInterval = config.CorrectionInterval
	}

	if config.Watchdog != nil {
		setup.Watchdog = config.Watchdog
	}

	if config.SlopeStartManual != nil {
		setup.SlopeStartManual = config.SlopeStartManual
	}

	if config.SlopeStopManual != nil {
		setup.SlopeStopManual = config.SlopeStopManual
	}

	if config.SlopeStartAuto != nil {
		setup.SlopeStartAuto = config.SlopeStartAuto
	}

	if config.SlopeStopAuto != nil {
		setup.SlopeStopAuto = config.SlopeStopAuto
	}

	if config.SensorRule != nil {
		setup.SensorRule = config.SensorRule
	}

	if config.Auto != nil {
		setup.Auto = config.Auto
	}

	if config.RuleBrightness != nil {
		setup.RuleBrightness = config.RuleBrightness
	}

	if config.RulePresence != nil {
		setup.RulePresence = config.RulePresence
	}

	if config.FirstDay != nil {
		setup.FirstDay = config.FirstDay
	}

	if config.FirstDayOffset != nil {
		setup.FirstDayOffset = config.FirstDayOffset
	}

	if config.SetpointOccupiedCool1 != nil {
		setup.SetpointOccupiedCool1 = config.SetpointOccupiedCool1
	}

	if config.SetpointOccupiedHeat1 != nil {
		setup.SetpointOccupiedHeat1 = config.SetpointOccupiedHeat1
	}

	if config.SetpointUnoccupiedCool1 != nil {
		setup.SetpointUnoccupiedCool1 = config.SetpointUnoccupiedCool1
	}

	if config.SetpointUnoccupiedHeat1 != nil {
		setup.SetpointUnoccupiedHeat1 = config.SetpointUnoccupiedHeat1
	}

	if config.SetpointStandbyCool1 != nil {
		setup.SetpointStandbyCool1 = config.SetpointStandbyCool1
	}

	if config.SetpointStandbyHeat1 != nil {
		setup.SetpointStandbyHeat1 = config.SetpointStandbyHeat1
	}

	if config.HvacsTargetMode != nil {
		setup.HvacsTargetMode = config.HvacsTargetMode
	}

	if config.HvacsHeatCool != nil {
		setup.HvacsHeatCool = config.HvacsHeatCool
	}

	return db.UpdateRecord(pconst.DbConfig, pconst.TbGroups, dbID, setup)
}

//GetGroupConfigs get group Config
func GetGroupConfigs(db Database, driversMac map[string]bool) map[int]gm.GroupConfig {
	groups := make(map[int]gm.GroupConfig)
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbGroups)
	if err != nil || stored == nil {
		return groups
	}
	for _, val := range stored {
		gr, err := gm.ToGroupConfig(val)
		if err != nil || gr == nil {
			continue
		}
		addGroup := false
		for _, mac := range gr.Leds {
			if _, ok := driversMac[mac]; ok {
				addGroup = true
				break
			}
		}
		if addGroup != true {
			for _, mac := range gr.Blinds {
				if _, ok := driversMac[mac]; ok {
					addGroup = true
					break
				}
			}
		}
		if addGroup != true {
			for _, mac := range gr.Hvacs {
				if _, ok := driversMac[mac]; ok {
					addGroup = true
					break
				}
			}
		}
		if addGroup {
			groups[gr.Group] = *gr
		}
	}
	return groups
}

//SaveGroupStatus dump group status in database
func SaveGroupStatus(db Database, status gm.GroupStatus) error {
	criteria := make(map[string]interface{})
	criteria["Group"] = status.Group
	return SaveOnUpdateObject(db, status, pconst.DbStatus, pconst.TbGroups, criteria)
}

//GetGroupStatus return the group status
func GetGroupStatus(db Database, grID int) *gm.GroupStatus {
	criteria := make(map[string]interface{})
	criteria["Group"] = grID
	stored, err := db.GetRecord(pconst.DbStatus, pconst.TbGroups, criteria)
	if err != nil || stored == nil {
		return nil
	}
	gr, err := gm.ToGroupStatus(stored)
	if err != nil {
		return nil
	}
	return gr
}

//GetGroupsStatus get groups status
func GetGroupsStatus(db Database) map[int]gm.GroupStatus {
	groups := make(map[int]gm.GroupStatus)
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbGroups)
	if err != nil || stored == nil {
		return groups
	}
	for _, val := range stored {
		gr, err := gm.ToGroupStatus(val)
		if err != nil || gr == nil {
			continue
		}
		groups[gr.Group] = *gr
	}
	return groups
}
