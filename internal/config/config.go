package config

import (
	"fmt"

	"github.com/jinzhu/configor"
)

var Config = struct {
	App struct {
		Debug bool
	}
	Server struct {
		Port uint64
	}
	Wechat struct {
		AppId     string
		AppSecret string
		Token     string
		AESKey    string
	}
}{}

// initConfig loads configuration file.
func init() {
	err := configor.Load(&Config, "config.yml")

	if err != nil {
		fmt.Println(err)
	}
}
