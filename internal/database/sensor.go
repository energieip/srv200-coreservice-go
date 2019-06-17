package database

import (
	ds "github.com/energieip/common-components-go/pkg/dsensor"
)

//SaveSensorConfig dump sensor config in database
func SaveSensorConfig(db Database, cfg ds.SensorSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	return SaveOnUpdateObject(db, cfg, ConfigDB, SensorsTable, criteria)
}

//SaveSensorLabelConfig dump sensor config in database
func SaveSensorLabelConfig(db Database, cfg ds.SensorSetup) error {
	criteria := make(map[string]interface{})
	if cfg.Label == nil {
		return NewError("Device " + cfg.Mac + "not found")
	}
	criteria["Label"] = *cfg.Label
	return SaveOnUpdateObject(db, cfg, ConfigDB, SensorsTable, criteria)
}

//UpdateSensorConfig update sensor config in database
func UpdateSensorConfig(db Database, cfg ds.SensorConf) error {
	setup, dbID := GetSensorConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		return NewError("Device " + cfg.Mac + "not found")
	}

	if cfg.BrightnessCorrectionFactor != nil {
		setup.BrightnessCorrectionFactor = cfg.BrightnessCorrectionFactor
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

	if cfg.TemperatureOffset != nil {
		setup.TemperatureOffset = cfg.TemperatureOffset
	}

	if cfg.ThresholdPresence != nil {
		setup.ThresholdPresence = cfg.ThresholdPresence
	}

	if cfg.DumpFrequency != nil {
		setup.DumpFrequency = *cfg.DumpFrequency
	}

	if cfg.IBeaconMajor != nil {
		setup.IBeaconMajor = cfg.IBeaconMajor
	}

	if cfg.IBeaconMinor != nil {
		setup.IBeaconMinor = cfg.IBeaconMinor
	}

	if cfg.IBeaconTxPower != nil {
		setup.IBeaconTxPower = cfg.IBeaconTxPower
	}

	if cfg.IBeaconUUID != nil {
		setup.IBeaconUUID = cfg.IBeaconUUID
	}

	if cfg.BleMode != nil {
		setup.BleMode = cfg.BleMode
	}

	return db.UpdateRecord(ConfigDB, SensorsTable, dbID, setup)
}

//UpdateSensorLabelSetup update sensor config in database
func UpdateSensorLabelSetup(db Database, cfg ds.SensorSetup) error {
	setup, dbID := GetSensorLabelConfig(db, *cfg.Label)
	if setup == nil || dbID == "" {
		if cfg.BleMode == nil {
			ble := "service"
			cfg.BleMode = &ble
		}
		if cfg.IsBleEnabled == nil {
			bleEnable := false
			cfg.IsBleEnabled = &bleEnable
		}
		if cfg.Group == nil {
			group := 0
			cfg.Group = &group
		}
		if cfg.FriendlyName == nil {
			name := *cfg.Label
			cfg.FriendlyName = &name
		}
		defaultValue := 0
		if cfg.BrightnessCorrectionFactor == nil {
			cfg.BrightnessCorrectionFactor = &defaultValue
		}
		if cfg.DumpFrequency == 0 {
			cfg.DumpFrequency = 1000
		}
		if cfg.TemperatureOffset == nil {
			cfg.TemperatureOffset = &defaultValue
		}
		if cfg.ThresholdPresence == nil {
			presence := 10
			cfg.ThresholdPresence = &presence
		}

		return SaveSensorLabelConfig(db, cfg)
	}

	if cfg.BrightnessCorrectionFactor != nil {
		setup.BrightnessCorrectionFactor = cfg.BrightnessCorrectionFactor
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

	if cfg.TemperatureOffset != nil {
		setup.TemperatureOffset = cfg.TemperatureOffset
	}

	if cfg.ThresholdPresence != nil {
		setup.ThresholdPresence = cfg.ThresholdPresence
	}

	if cfg.DumpFrequency != 0 {
		setup.DumpFrequency = cfg.DumpFrequency
	}

	if cfg.Label != nil {
		setup.Label = cfg.Label
	}

	if cfg.IBeaconMajor != nil {
		setup.IBeaconMajor = cfg.IBeaconMajor
	}

	if cfg.IBeaconMinor != nil {
		setup.IBeaconMinor = cfg.IBeaconMinor
	}

	if cfg.IBeaconTxPower != nil {
		setup.IBeaconTxPower = cfg.IBeaconTxPower
	}

	if cfg.IBeaconUUID != nil {
		setup.IBeaconUUID = cfg.IBeaconUUID
	}

	if cfg.BleMode != nil {
		setup.BleMode = cfg.BleMode
	}
	return db.UpdateRecord(ConfigDB, SensorsTable, dbID, setup)
}

//RemoveSensorConfig remove sensor config in database
func RemoveSensorConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(ConfigDB, SensorsTable, criteria)
}

//RemoveSensorStatus remove led status in database
func RemoveSensorStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(StatusDB, SensorsTable, criteria)
}

//GetSensorSwitchStatus get cluster Config list
func GetSensorSwitchStatus(db Database, swMac string) map[string]ds.Sensor {
	res := map[string]ds.Sensor{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(StatusDB, SensorsTable, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := ds.ToSensor(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetSensorSwitchSetup get sensor Config list
func GetSensorSwitchSetup(db Database, swMac string) map[string]ds.SensorSetup {
	res := map[string]ds.SensorSetup{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(StatusDB, SensorsTable, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := ds.ToSensorSetup(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetSensorConfig return the sensor configuration
func GetSensorConfig(db Database, mac string) (*ds.SensorSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(ConfigDB, SensorsTable, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := ds.ToSensorSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//GetSensorLabelConfig return the sensor configuration
func GetSensorLabelConfig(db Database, label string) (*ds.SensorSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(ConfigDB, SensorsTable, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := ds.ToSensorSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//SwitchSensorConfig update sensor config in database
func SwitchSensorConfig(db Database, old, oldFull, new, newFull string) error {
	setup, dbID := GetSensorConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + "not found")
	}
	setup.FullMac = newFull
	setup.Mac = new
	return db.UpdateRecord(ConfigDB, SensorsTable, dbID, setup)
}

//GetSensorsConfig return the sensor config list
func GetSensorsConfig(db Database) map[string]ds.SensorSetup {
	drivers := map[string]ds.SensorSetup{}
	stored, err := db.FetchAllRecords(ConfigDB, SensorsTable)
	if err != nil || stored == nil {
		return drivers
	}
	for _, l := range stored {
		driver, err := ds.ToSensorSetup(l)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//SaveSensorStatus dump sensor status in database
func SaveSensorStatus(db Database, status ds.Sensor) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, status, StatusDB, SensorsTable, criteria)
}

//GetSensorsStatus return the led status list
func GetSensorsStatus(db Database) map[string]ds.Sensor {
	drivers := map[string]ds.Sensor{}
	stored, err := db.FetchAllRecords(StatusDB, SensorsTable)
	if err != nil || stored == nil {
		return drivers
	}
	for _, l := range stored {
		driver, err := ds.ToSensor(l)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//GetSensorStatus return the led status
func GetSensorStatus(db Database, mac string) *ds.Sensor {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(StatusDB, SensorsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	driver, err := ds.ToSensor(stored)
	if err != nil {
		return nil
	}
	return driver
}
