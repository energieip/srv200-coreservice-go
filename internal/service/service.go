package service

import (
	"os"

	"github.com/energieip/common-group-go/pkg/groupmodel"
	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
	pkg "github.com/energieip/common-service-go/pkg/service"
	"github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/srv200-coreservice-go/internal/api"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/energieip/srv200-coreservice-go/internal/network"
	"github.com/romana/rlog"
)

const (
	ActionReload = "reload"
	ActionSetup  = "setup"
	ActionDump   = "dump"
	ActionRemove = "remove"

	UrlStatus = "status/dump"
	UrlHello  = "setup/hello"

	EventRemove = "remove"
	EventUpdate = "update"
	EventAdd    = "add"
)

//CoreService content
type CoreService struct {
	server          network.ServerNetwork //Remote server
	db              database.Database
	mac             string //Switch mac address
	events          chan string
	installMode     bool
	eventsAPI       chan map[string]interface{}
	eventsToBackend chan map[string]interface{}
	api             *api.API
}

//Initialize service
func (s *CoreService) Initialize(confFile string) error {
	clientID := "Server"
	s.installMode = false
	s.events = make(chan string)
	s.eventsAPI = make(chan map[string]interface{})

	conf, err := pkg.ReadServiceConfig(confFile)
	if err != nil {
		rlog.Error("Cannot parse configuration file " + err.Error())
		return err
	}
	os.Setenv("RLOG_LOG_LEVEL", conf.LogLevel)
	os.Setenv("RLOG_LOG_NOTIME", "yes")
	rlog.UpdateEnv()
	rlog.Info("Starting ServerCore service")

	db, err := database.ConnectDatabase(conf.DB.ClientIP, conf.DB.ClientPort)
	if err != nil {
		rlog.Error("Cannot connect to database " + err.Error())
		return err
	}
	s.db = *db

	serverNet, err := network.CreateServerNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.NetworkBroker.IP + " error: " + err.Error())
		return err
	}
	s.server = *serverNet

	err = s.server.LocalConnection(*conf, clientID)
	if err != nil {
		rlog.Error("Cannot connect to drivers broker " + conf.NetworkBroker.IP + " error: " + err.Error())
		return err
	}
	web := api.InitAPI(s.db, s.eventsAPI)
	s.api = web

	rlog.Info("ServerCore service started")
	return nil
}

//Stop service
func (s *CoreService) Stop() {
	rlog.Info("Stopping ServerCore service")
	s.server.Disconnect()
	s.db.Close()
	rlog.Info("ServerCore service stopped")
}

func (s *CoreService) prepareSetupSwitchConfig(switchStatus deviceswitch.SwitchStatus) *deviceswitch.SwitchConfig {
	config := database.GetSwitchConfig(s.db, switchStatus.Mac)
	if config == nil && !s.installMode {
		return nil
	}

	isConfigured := true
	setup := deviceswitch.SwitchConfig{}
	setup.Mac = switchStatus.Mac
	setup.FriendlyName = config.FriendlyName
	setup.IsConfigured = &isConfigured
	setup.Services = database.GetServiceConfigs(s.db)
	if s.installMode {
		switchSetup := core.SwitchSetup{}
		switchSetup.Mac = setup.Mac
		switchSetup.IP = switchStatus.IP
		switchSetup.Cluster = 0
		switchSetup.FriendlyName = switchStatus.FriendlyName
		database.SaveSwitchConfig(s.db, switchSetup)
	}
	if config.IP == "" {
		config.IP = switchStatus.IP
		database.SaveSwitchConfig(s.db, *config)
	}
	return &setup
}

func (s *CoreService) prepareSwitchConfig(switchStatus deviceswitch.SwitchStatus) *deviceswitch.SwitchConfig {
	config := database.GetSwitchConfig(s.db, switchStatus.Mac)
	if config == nil && !s.installMode {
		return nil
	}

	isConfigured := true
	setup := deviceswitch.SwitchConfig{}
	setup.Mac = switchStatus.Mac
	setup.FriendlyName = config.FriendlyName
	setup.IsConfigured = &isConfigured

	defaultGroup := 0
	defaultWatchdog := 600

	setup.LedsSetup = make(map[string]driverled.LedSetup)
	setup.SensorsSetup = make(map[string]driversensor.SensorSetup)

	driversMac := make(map[string]bool)
	for _, led := range switchStatus.Leds {
		driversMac[led.Mac] = true
	}
	setup.Groups = database.GetGroupConfigs(s.db, driversMac)

	for mac, led := range switchStatus.Leds {
		if !led.IsConfigured {
			lsetup := database.GetLedConfig(s.db, mac)
			if lsetup == nil && s.installMode {
				enableBle := false
				low := 0
				high := 100
				dled := driverled.LedSetup{
					Mac:          led.Mac,
					IMax:         100,
					Group:        &defaultGroup,
					Watchdog:     &defaultWatchdog,
					IsBleEnabled: &enableBle,
					ThresoldHigh: &high,
					ThresoldLow:  &low,
				}
				lsetup = &dled
				// saved default config
				database.SaveLedConfig(s.db, dled)
			}
			if lsetup != nil {
				setup.LedsSetup[mac] = *lsetup
				event := make(map[string]interface{})
				evtType := "led." + EventAdd
				event[evtType] = led
				s.eventsAPI <- event
			}
		} else {
			event := make(map[string]interface{})
			evtType := "led." + EventUpdate
			event[evtType] = led
			s.eventsAPI <- event
		}
	}

	for mac, sensor := range switchStatus.Sensors {
		if !sensor.IsConfigured {
			ssetup := database.GetSensorConfig(s.db, mac)
			if ssetup == nil {
				enableBle := true
				brightnessCorrection := 1
				thresoldPresence := 10
				temperatureOffset := 0
				dsensor := driversensor.SensorSetup{
					Mac:                        sensor.Mac,
					Group:                      &defaultGroup,
					IsBleEnabled:               &enableBle,
					BrigthnessCorrectionFactor: &brightnessCorrection,
					ThresoldPresence:           &thresoldPresence,
					TemperatureOffset:          &temperatureOffset,
				}
				ssetup = &dsensor
				// saved default config
				database.SaveSensorConfig(s.db, dsensor)
			}
			if ssetup != nil {
				setup.SensorsSetup[mac] = *ssetup
				event := make(map[string]interface{})
				evtType := "sensor." + EventAdd
				event[evtType] = sensor
				s.eventsAPI <- event
			}
		} else {
			event := make(map[string]interface{})
			evtType := "sensor." + EventUpdate
			event[evtType] = sensor
			s.eventsAPI <- event
		}
	}
	return &setup
}

func (s *CoreService) sendSwitchSetup(sw deviceswitch.SwitchStatus) {
	conf := s.prepareSetupSwitchConfig(sw)
	if conf == nil {
		rlog.Warn("This device " + sw.Mac + " is not authorized")
		return
	}
	switchSetup := *conf

	url := "/write/" + sw.Topic + "/setup/config"
	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send setup config to " + sw.Mac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + sw.Mac + " on topic: " + url)
	}
}

func (s *CoreService) sendSwitchUpdateConfig(sw deviceswitch.SwitchStatus) {
	conf := s.prepareSwitchConfig(sw)
	if conf == nil {
		rlog.Warn("This device " + sw.Mac + " is not authorized")
		return
	}
	switchSetup := *conf

	url := "/write/" + sw.Topic + "/update/settings"
	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + sw.Mac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + sw.Mac + " on topic: " + url)
	}
}

func (s *CoreService) sendSwitchCommand(cmd core.ServerCmd) {
	rlog.Info("Send switch cmd", cmd)
}

func (s *CoreService) registerSwitchStatus(switchStatus deviceswitch.SwitchStatus) {
	for _, led := range switchStatus.Leds {
		database.SaveLedStatus(s.db, led)
	}
	for _, sensor := range switchStatus.Sensors {
		database.SaveSensorStatus(s.db, sensor)
	}
	for _, group := range switchStatus.Groups {
		database.SaveGroupStatus(s.db, group)
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

func (s *CoreService) registerConfig(config core.ServerConfig) {
	database.SaveServerConfig(s.db, config)
}

func (s *CoreService) updateLedCfg(config interface{}) {
	cfg, _ := driverled.ToLedConf(config)
	database.UpdateLedConfig(s.db, *cfg)
	//TODO send order to switch
}

func (s *CoreService) updateGroupCfg(config interface{}) {
	cfg, _ := groupmodel.ToGroupConfig(config)
	database.UpdateGroupConfig(s.db, *cfg)
	//TODO send order to switch
}

func (s *CoreService) updateSwitchCfg(config interface{}) {
	cfg, _ := core.ToSwitchConfig(config)
	database.UpdateSwitchConfig(s.db, *cfg)
	//TODO send order to switch
}

func (s *CoreService) updateSensorCfg(config interface{}) {
	cfg, _ := driversensor.ToSensorConf(config)
	database.UpdateSensorConfig(s.db, *cfg)
	//TODO send order to switch
}

func (s *CoreService) sendGroupCmd(cmd interface{}) {
	rlog.Info("TODO send command to group", cmd)
}

func (s *CoreService) sendLedCmd(cmd interface{}) {
	rlog.Info("TODO send command to LED", cmd)
}

func (s *CoreService) readAPIEvents() {
	for {
		select {
		case apiEvents := <-s.api.EventsToBackend:
			for eventType, event := range apiEvents {
				switch eventType {
				case "led":
					s.updateLedCfg(event)
				case "sensor":
					s.updateSensorCfg(event)
				case "group":
					s.updateGroupCfg(event)
				case "switch":
					s.updateSwitchCfg(event)
				case "groupCmd":
					s.sendGroupCmd(event)
				case "ledCmd":
					s.sendLedCmd(event)
				}
			}
		}
	}
}

//Run service mainloop
func (s *CoreService) Run() error {
	go s.readAPIEvents()
	for {
		select {
		case serverEvents := <-s.server.Events:
			for eventType, event := range serverEvents {
				switch eventType {
				case network.EventHello:
					s.sendSwitchSetup(event)
					s.registerSwitchStatus(event)
				case network.EventDump:
					s.sendSwitchUpdateConfig(event)
					s.registerSwitchStatus(event)
				}
			}
		case configEvents := <-s.server.EventsCfg:
			for eventType, event := range configEvents {
				switch eventType {
				case network.EventWriteCfg:
					s.registerConfig(event)
				}
			}
		case installModeEvent := <-s.server.EventsInstallMode:
			s.installMode = installModeEvent
			rlog.Info("Installation mode status is now:", s.installMode)
		case cmd := <-s.server.EventsCmd:
			s.sendSwitchCommand(cmd)
		}
	}
}
