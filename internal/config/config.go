package config

import (
	"flag"

	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"
)

type Args struct {
	Address       string
	DatabaseURL   string
	SystemAddress string
}

type Param struct {
	Address       string `env:"RUN_ADDRESS"`
	DatabaseURL   string `env:"DATABASE_URI"`
	SystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func loadConfig() (p Param, err error) {
	var parametrs Param
	err = env.Parse(&parametrs)
	if err != nil {
		logrus.Error("Error parse env: ", err)
	}
	return parametrs, nil
}

var Arg Args

func TerminateFlags() (err error) {
	env, err := loadConfig()
	if err != nil {
		logrus.Error("Error terminate env or flags: ", err)
	}
	var Flags Args
	flag.StringVar(&Flags.Address, "a", "127.0.0.1:8080", "Address of server")
	flag.StringVar(&Flags.DatabaseURL, "d", "postgres://user:user@127.0.0.1:5432/db", "Database URL")
	flag.StringVar(&Flags.SystemAddress, "r", "127.0.0.1:8090", "Address of system accrual")
	flag.Parse()
	logrus.Info("[[[[[[[[[[[[[[[", env)
	logrus.Info("5555555555555555555555555555[[[[[[[[[[[[[[[", &Flags)

	if env.Address == "" {
		Arg.Address = Flags.Address
	} else {
		Arg.Address = env.Address
	}
	if env.DatabaseURL == "" {
		Arg.DatabaseURL = Flags.DatabaseURL
	} else {
		Arg.DatabaseURL = env.DatabaseURL
	}
	if env.SystemAddress == "" {
		Arg.SystemAddress = Flags.SystemAddress
	} else {
		Arg.SystemAddress = env.SystemAddress
	}
	return err
}
