package config

import (
	"flag"
	"os"

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
	flag.StringVar(&Flags.SystemAddress, "r", "http://127.0.0.1:8090", "Address of system accrual")
	flag.Parse()
	logrus.Info("[[[[[[[[[[[[[[[ ENV", env)
	logrus.Info("5555555555555555555555555555[[[[[[[[[[[[[[[ FLAGS", &Flags)
	envAddress, _ := os.LookupEnv("RUN_ADDRESS")
	if envAddress == "" {
		Arg.Address = Flags.Address
	} else {
		Arg.Address = env.Address
	}
	envURI, _ := os.LookupEnv("DATABASE_URI")
	if envURI == "" {
		Arg.DatabaseURL = Flags.DatabaseURL
	} else {
		Arg.DatabaseURL = env.DatabaseURL
	}
	envSystemAddress, _ := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if envSystemAddress == "" {
		Arg.SystemAddress = Flags.SystemAddress
	} else {
		Arg.SystemAddress = env.SystemAddress
	}
	return err
}
