package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type config struct {
	SSID       string // ssid of wifi network
	Password   string // password of wifi network
	LanGw      string // lan network gateway
	WifiGw     string // wifi network gateway
	LanIef     string // lan interface name
	WifiIef    string // wifi interface name
	Server     string // remote server url
	ListenHost string // server listen host
	ListenPort int    // server listen port
}

var C config
var err error

func InitConfig() error {
	_, err = toml.DecodeFile("./config.toml", &C)
	if err != nil {
		return fmt.Errorf("Failed to decode config: %s", err.Error())
	}
	return nil
}
