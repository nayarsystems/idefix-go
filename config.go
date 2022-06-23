package idefixgo

import (
	"runtime/debug"

	"github.com/spf13/viper"
)

func readConfig(name string) (*ClientOptions, error) {
	c := &ClientOptions{}

	viper.SetConfigName(name)

	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.idefix/")
	viper.AddConfigPath("$HOME/.config/idefix/")

	viper.SetDefault("BrokerAddress", "ssl://mqtt.terathings.com:8883")
	viper.SetDefault("Encoding", "mg")

	info, _ := debug.ReadBuildInfo()

	viper.SetDefault("Meta", map[string]interface{}{"idefix-go": info.Main.Version})

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(c); err != nil {
		return nil, err
	}

	return c, nil
}
