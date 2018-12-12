package database

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveModel dump model in database
func SaveModel(db Database, m core.Model) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Name"] = m.Name
	stored, err := db.GetRecord(ConfigDB, ModelsTable, criteria)
	if err == nil && stored != nil {
		m := stored.(map[string]interface{})
		id, ok := m["id"]
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, ModelsTable, m)
	} else {
		err = db.UpdateRecord(ConfigDB, ModelsTable, dbID, m)
	}
	return err
}

//GetModel return the led configuration
func GetModel(db Database, name string) *core.Model {
	criteria := make(map[string]interface{})
	criteria["Name"] = name
	stored, err := db.GetRecord(ConfigDB, ModelsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	model, err := core.ToModel(stored)
	if err != nil {
		return nil
	}
	return model
}
