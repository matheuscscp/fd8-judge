package application

import (
	"github.com/spf13/viper"
)

const CONFIG_KEY_APP_TYPE = "app_type"

type App interface {
	Run()
	Shutdown()
}

func SetupConfigs() {
	viper.SetDefault(CONFIG_KEY_APP_TYPE, CliAppType)
	viper.BindEnv(CONFIG_KEY_APP_TYPE)
}

func GetAppType() string {
	return viper.Get(CONFIG_KEY_APP_TYPE).(string)
}
