package service

import (
	"github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/energieip/srv200-coreservice-go/internal/history"
	"github.com/romana/rlog"
)

func (s *CoreService) updateSwitchCfg(config interface{}) {
	cfg, _ := core.ToSwitchConfig(config)
	sw := database.GetSwitchConfig(s.db, cfg.Mac)
	if sw != nil {
		database.UpdateSwitchConfig(s.db, *cfg)
	} else {
		database.SaveSwitchConfig(s.db, *cfg)
	}

	url := "/write/switch/" + cfg.Mac + "/update/settings"
	switchCfg := sd.SwitchConfig{}
	switchCfg.Mac = cfg.Mac
	if cfg.DumpFrequency != nil {
		switchCfg.DumpFrequency = *cfg.DumpFrequency
	}
	switchCfg.FriendlyName = cfg.FriendlyName
	if cfg.IsConfigured != nil {
		switchCfg.IsConfigured = cfg.IsConfigured
	}

	dump, _ := switchCfg.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) registerSwitchStatus(switchStatus sd.SwitchStatus) {
	oldLeds := database.GetLedSwitchStatus(s.db, switchStatus.Mac)
	for mac, led := range switchStatus.Leds {
		database.SaveLedStatus(s.db, led)
		_, ok := oldLeds[mac]
		if ok {
			delete(oldLeds, mac)
		}
	}
	for _, led := range oldLeds {
		database.RemoveLedStatus(s.db, led.Mac)
		s.prepareAPIEvent(EventRemove, LedElt, led)
	}

	oldSensors := database.GetSensorSwitchStatus(s.db, switchStatus.Mac)
	for mac, sensor := range switchStatus.Sensors {
		database.SaveSensorStatus(s.db, sensor)
		_, ok := oldSensors[mac]
		if ok {
			delete(oldSensors, mac)
		}
	}
	for _, sensor := range oldSensors {
		database.RemoveSensorStatus(s.db, sensor.Mac)
		s.prepareAPIEvent(EventRemove, SensorElt, sensor)
	}

	oldBlinds := database.GetBlindSwitchStatus(s.db, switchStatus.Mac)
	for mac, blind := range switchStatus.Blinds {
		database.SaveBlindStatus(s.db, blind)
		_, ok := oldBlinds[mac]
		if ok {
			delete(oldBlinds, mac)
		}
	}
	for _, blind := range oldBlinds {
		database.RemoveBlindStatus(s.db, blind.Mac)
		s.prepareAPIEvent(EventRemove, BlindElt, blind)
	}

	oldHvacs := database.GetHvacSwitchStatus(s.db, switchStatus.Mac)
	for mac, hvac := range switchStatus.Hvacs {
		database.SaveHvacStatus(s.db, hvac)
		_, ok := oldHvacs[mac]
		if ok {
			delete(oldHvacs, mac)
		}
	}
	for _, hvac := range oldHvacs {
		database.RemoveHvacStatus(s.db, hvac.Mac)
		s.prepareAPIEvent(EventRemove, HvacElt, hvac)
	}

	for _, group := range switchStatus.Groups {
		database.SaveGroupStatus(s.db, group)
		s.prepareAPIEvent(EventUpdate, GroupElt, group)
	}

	for _, service := range switchStatus.Services {
		serv := core.ServiceDump{}
		serv.Name = service.Name
		serv.PackageName = service.PackageName
		serv.Version = service.Version
		serv.Status = service.Status
		serv.SwitchMac = switchStatus.Mac
		database.SaveServiceStatus(s.db, serv)
	}
	database.SaveSwitchStatus(s.db, switchStatus)
}

func (s *CoreService) sendSwitchSetup(sw sd.SwitchStatus) {
	conf := s.prepareSetupSwitchConfig(sw)
	if conf == nil {
		rlog.Warn("This device " + sw.Mac + " is not authorized")
		return
	}
	switchSetup := *conf

	url := "/write/switch/" + sw.Mac + "/setup/config"
	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) sendSwitchRemoveConfig(sw sd.SwitchConfig) {
	url := "/remove/switch/" + sw.Mac + "/update/settings"
	dump, _ := sw.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) sendSwitchUpdateConfig(sw sd.SwitchStatus) {
	conf := s.prepareSwitchConfig(sw)
	if conf == nil {
		rlog.Warn("This device " + sw.Mac + " is not authorized")
		return
	}
	switchSetup := *conf

	url := "/write/switch/" + sw.Mac + "/update/settings"
	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) prepareSetupSwitchConfig(switchStatus sd.SwitchStatus) *sd.SwitchConfig {
	config := database.GetSwitchConfig(s.db, switchStatus.Mac)
	if config == nil {
		return nil
	}

	isConfigured := true
	setup := sd.SwitchConfig{}
	setup.Mac = switchStatus.Mac
	setup.FriendlyName = config.FriendlyName
	setup.IsConfigured = &isConfigured
	setup.LedsSetup = database.GetLedSwitchSetup(s.db, switchStatus.Mac)
	setup.SensorsSetup = database.GetSensorSwitchSetup(s.db, switchStatus.Mac)
	setup.BlindsSetup = database.GetBlindSwitchSetup(s.db, switchStatus.Mac)
	setup.HvacsSetup = database.GetHvacSwitchSetup(s.db, switchStatus.Mac)
	newGroups := make(map[int]bool)

	driversMac := make(map[string]bool)
	for mac := range setup.LedsSetup {
		driversMac[mac] = true
	}
	for mac := range setup.SensorsSetup {
		driversMac[mac] = true
	}
	for mac := range setup.BlindsSetup {
		driversMac[mac] = true
	}
	for mac := range setup.HvacsSetup {
		driversMac[mac] = true
	}

	setup.Groups = database.GetGroupConfigs(s.db, driversMac)
	for _, gr := range setup.Groups {
		newGroups[gr.Group] = true
	}
	setup.Users = database.GetUserConfigs(s.db, newGroups, true)

	services := make(map[string]pkg.Service)
	srv := database.GetServiceConfigs(s.db)
	for _, service := range srv {
		val, ok := switchStatus.Services[service.Name]
		if !ok || val.Version != service.Version {
			services[service.Name] = service
		}
	}

	setup.Services = services
	if config.IP == "" {
		config.IP = switchStatus.IP
		database.SaveSwitchConfig(s.db, *config)
	}

	//Prepare Cluster
	var clusters map[string]core.SwitchConfig
	switchCluster := make(map[string]sd.SwitchCluster)
	if config.Cluster != 0 {
		clusters = database.GetCluster(s.db, config.Cluster)
	}
	for _, cluster := range clusters {
		if cluster.Mac != switchStatus.Mac {
			br := sd.SwitchCluster{
				IP:  cluster.IP,
				Mac: cluster.Mac,
			}
			switchCluster[cluster.Mac] = br
		}
	}
	setup.ClusterBroker = switchCluster
	return &setup
}

func (s *CoreService) prepareSwitchConfig(switchStatus sd.SwitchStatus) *sd.SwitchConfig {
	config := database.GetSwitchConfig(s.db, switchStatus.Mac)
	if config == nil {
		rlog.Warn("Cannot find configuration for switch", switchStatus.Mac)
		return nil
	}
	if config.IP == "" {
		config.IP = switchStatus.IP
		database.SaveSwitchConfig(s.db, *config)
	}

	isConfigured := true
	setup := sd.SwitchConfig{}
	setup.Mac = switchStatus.Mac
	setup.IP = config.IP
	setup.FriendlyName = config.FriendlyName
	setup.IsConfigured = &isConfigured

	setup.LedsSetup = make(map[string]dl.LedSetup)
	setup.SensorsSetup = make(map[string]ds.SensorSetup)
	setup.BlindsSetup = make(map[string]dblind.BlindSetup)
	setup.HvacsSetup = make(map[string]dhvac.HvacSetup)
	grList := make(map[int]bool)

	driversMac := make(map[string]bool)
	for _, led := range switchStatus.Leds {
		driversMac[led.Mac] = true
	}
	for _, blind := range switchStatus.Blinds {
		driversMac[blind.Mac] = true
	}
	for _, hvac := range switchStatus.Hvacs {
		driversMac[hvac.Mac] = true
	}
	newGroups := make(map[int]gm.GroupConfig)
	groups := database.GetGroupConfigs(s.db, driversMac)
	for _, gr := range groups {
		old, ok := switchStatus.Groups[gr.Group]
		if ok && !s.isGroupRequiredUpdate(old, gr) {
			continue
		}
		newGroups[gr.Group] = gr
		grList[gr.Group] = true
	}
	setup.Groups = newGroups
	setup.Users = database.GetUserConfigs(s.db, grList, true)

	for mac, led := range switchStatus.Leds {
		if !led.IsConfigured {
			lsetup, _ := database.GetLedConfig(s.db, mac)
			if lsetup != nil {
				setup.LedsSetup[mac] = *lsetup
			}
			s.prepareAPIEvent(EventAdd, LedElt, led)
		} else {
			s.prepareAPIEvent(EventUpdate, LedElt, led)
			history.SaveLedHistory(s.historyDb, led)
			s.prepareAPIConsumption(LedElt, led.LinePower)
		}
	}

	for mac, blind := range switchStatus.Blinds {
		if !blind.IsConfigured {
			bsetup, _ := database.GetBlindConfig(s.db, mac)
			if bsetup != nil {
				setup.BlindsSetup[mac] = *bsetup
			}
			s.prepareAPIEvent(EventAdd, BlindElt, blind)
		} else {
			s.prepareAPIEvent(EventUpdate, BlindElt, blind)
			history.SaveBlindHistory(s.historyDb, blind)
			s.prepareAPIConsumption(BlindElt, blind.LinePower)
		}
	}

	for mac, hvac := range switchStatus.Hvacs {
		if !hvac.IsConfigured {
			bsetup, _ := database.GetHvacConfig(s.db, mac)
			if bsetup != nil {
				setup.HvacsSetup[mac] = *bsetup
			}
			s.prepareAPIEvent(EventAdd, HvacElt, hvac)
		} else {
			s.prepareAPIEvent(EventUpdate, HvacElt, hvac)
			// history.SaveHvacHistory(s.historyDb, hvac)
			// s.prepareAPIConsumption(HvacElt, hvac.LinePower)
		}
	}

	for mac, sensor := range switchStatus.Sensors {
		if !sensor.IsConfigured {
			ssetup, _ := database.GetSensorConfig(s.db, mac)
			if ssetup != nil {
				setup.SensorsSetup[mac] = *ssetup
			}
			s.prepareAPIEvent(EventAdd, SensorElt, sensor)
		} else {
			s.prepareAPIEvent(EventUpdate, SensorElt, sensor)
		}
	}

	//Prepare Cluster
	var clusters map[string]core.SwitchConfig
	switchCluster := make(map[string]sd.SwitchCluster)
	if config.Cluster != 0 {
		clusters = database.GetCluster(s.db, config.Cluster)
		for _, cluster := range clusters {
			_, ok := switchStatus.ClusterBroker[cluster.Mac]
			if !ok {
				//add only new cluster member only
				if cluster.Mac != switchStatus.Mac {
					br := sd.SwitchCluster{
						IP:  cluster.IP,
						Mac: cluster.Mac,
					}
					switchCluster[cluster.Mac] = br
				}
			}
		}
	}
	setup.ClusterBroker = switchCluster
	return &setup
}
