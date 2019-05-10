package config

import (
	"github.com/si9ma/KillOJ-common/asyncjob"
	"github.com/si9ma/KillOJ-common/kredis"
	"github.com/si9ma/KillOJ-common/mysql"
)

type Config struct {
	AsyncJob asyncjob.Config `yaml:"asyncJob"`
	Mysql    mysql.Config    `yaml:"mysql"`
	Redis    kredis.Config   `yaml:"redis"`
	App      AppConfig       `yaml:"app"`
}

type AppConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (a AppConfig) Addr() string {
	return a.Host + ":" + a.Port
}
