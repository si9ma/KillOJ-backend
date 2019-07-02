package asyncjob

import (
	"encoding/json"
	"strings"

	mconfig "github.com/RichardKnop/machinery/v1/config"
	"github.com/si9ma/KillOJ-common/constants"
)

// wrap machinery config
type Config struct {
	Broker        string `yaml:"broker"`
	DefaultQueue  string `yaml:"default_queue"`
	Exchange      string `yaml:"exchange"`
	ExchangeType  string `yaml:"exchange_type"`
	BindingKey    string `yaml:"binding_key"`
	PrefetchCount int    `yaml:"prefetch_count"`
}

func (c Config) String() string {
	res, _ := json.Marshal(c)
	return string(res)
}

func defaultConfig() Config {
	return Config{
		Broker:        "amqp://guest:guest@localhost:5672/",
		DefaultQueue:  strings.Join([]string{constants.ProjectName, "Queue"}, "_"),
		Exchange:      strings.Join([]string{constants.ProjectName, "Exchange"}, "_"),
		ExchangeType:  "direct",
		BindingKey:    strings.Join([]string{constants.ProjectName, "Queue"}, "_"),
		PrefetchCount: 3,
	}
}

// convert config to machinery config
func (c Config) toMachineryCfg() *mconfig.Config {
	return &mconfig.Config{
		Broker:       c.Broker,
		DefaultQueue: c.DefaultQueue,
		AMQP: &mconfig.AMQPConfig{
			Exchange:      c.Exchange,
			ExchangeType:  c.ExchangeType,
			BindingKey:    c.BindingKey,
			PrefetchCount: c.PrefetchCount,
		},
	}
}
