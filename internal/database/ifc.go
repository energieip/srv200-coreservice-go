package database

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

func GetIfcs(db Database) []core.IfcInfo {
	var res []core.IfcInfo
	projects := GetProjects(db)
	models := GetModels(db)

	for _, project := range projects {
		if project.ModelName == nil {
			continue
		}
		model, ok := models[*project.ModelName]
		if !ok {
			continue
		}
		mac := ""
		if project.Mac != nil {
			mac = *project.Mac
		}
		ifc := core.IfcInfo{
			Label:          project.Label,
			ModelName:      model.Name,
			Mac:            mac,
			Vendor:         model.Vendor,
			URL:            model.URL,
			DeviceType:     model.DeviceType,
			ProductionYear: model.ProductionYear,
		}
		if project.ModbusID != nil {
			ifc.ModbusID = project.ModbusID
		}
		if project.SlaveID != nil {
			ifc.SlaveID = project.SlaveID
		}
		res = append(res, ifc)
	}
	return res
}
