package env

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	NodeEnv = &NodeConfig{}
)

func parser() {
	v := viper.New()
	configFile := "./env.yaml"
	v.SetConfigFile(configFile)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(&NodeEnv); err != nil {
		panic(err)
	}
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		if err := v.ReadInConfig(); err != nil {
			fmt.Printf("OnConfigChange ReadInConfig err. %s", err.Error())
		}
		if err := v.Unmarshal(&NodeEnv); err != nil {
			fmt.Printf("OnConfigChange Unmarshal err. %s", err.Error())
		}
	})
}
