package asyncjob

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/brokers/iface"

	machlog "github.com/RichardKnop/machinery/v1/log"
)

// manual init
func Init(cfg Config) (machineryServer *machinery.Server, err error) {
	machineryCfg := cfg.toMachineryCfg()

	var machineryBroker iface.Broker
	if machineryBroker, err = machinery.BrokerFactory(machineryCfg); err != nil {
		return nil, err
	}
	machineryBackend := NewBlankBackend()

	machineryServer = machinery.NewServerWithBrokerBackend(machineryCfg, machineryBroker, machineryBackend)
	machlog.Set(ZLogger{})          // custom logger
	machlog.SetDebug(DebugLogger{}) // set debug logger
	return
}
