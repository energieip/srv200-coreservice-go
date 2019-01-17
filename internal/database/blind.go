package database

import (
	"github.com/energieip/common-components-go/pkg/dblind"
)

//SaveBlindConfig dump blind config in database
func SaveBlindConfig(db Database, cfg dblind.BlindSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	return SaveOnUpdateObject(db, cfg, ConfigDB, BlindsTable, criteria)
}

//UpdateBlindConfig update blind config in database
func UpdateBlindConfig(db Database, cfg dblind.BlindConf) error {
	setup, dbID := GetBlindConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		return NewError("Device " + cfg.Mac + "not found")
	}

	if cfg.FriendlyName != nil {
		setup.FriendlyName = cfg.FriendlyName
	}

	if cfg.Group != nil {
		setup.Group = cfg.Group
	}

	if cfg.IsBleEnabled != nil {
		setup.IsBleEnabled = cfg.IsBleEnabled
	}

	if cfg.DumpFrequency != nil {
		setup.DumpFrequency = *cfg.DumpFrequency
	}

	return db.UpdateRecord(ConfigDB, BlindsTable, dbID, setup)
}

//RemoveBlindConfig remove blind config in database
func RemoveBlindConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(ConfigDB, BlindsTable, criteria)
}

//GetBlindConfig return the sensor configuration
func GetBlindConfig(db Database, mac string) (*dblind.BlindSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(ConfigDB, BlindsTable, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := dblind.ToBlindSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//GetBlindsConfig return the blind config list
func GetBlindsConfig(db Database) map[string]dblind.BlindSetup {
	drivers := map[string]dblind.BlindSetup{}
	stored, err := db.FetchAllRecords(ConfigDB, BlindsTable)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dblind.ToBlindSetup(s)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//SaveBlindStatus dump blind status in database
func SaveBlindStatus(db Database, status dblind.Blind) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, status, StatusDB, BlindsTable, criteria)
}

//GetBlindsStatus return the blind status list
func GetBlindsStatus(db Database) map[string]dblind.Blind {
	drivers := map[string]dblind.Blind{}
	stored, err := db.FetchAllRecords(StatusDB, BlindsTable)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dblind.ToBlind(s)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//GetBlindStatus return the blind status
func GetBlindStatus(db Database, mac string) *dblind.Blind {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(StatusDB, BlindsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	driver, err := dblind.ToBlind(stored)
	if err != nil {
		return nil
	}
	return driver
}