package main

import (
	"github.com/BurntSushi/toml"
)

type config struct {
	RouteB   routeB `toml:"routeb"`
	Database database
	Log      logger `toml:"logger"`
}

type routeB struct {
	Id  string
	Pwd string `toml:"password"`
}

type database struct {
	Host string
	Port int
}

type logger struct {
	Level string
}

func loadConfig(path string, conf *config) error {
	_, err := toml.DecodeFile(path, conf)
	return err
}
