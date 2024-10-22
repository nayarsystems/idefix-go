package idefixgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/viper"
)

// ReadConfig function reads the config from the filename given and returns the [ClientOptions] struct with the config data added. File to read must be in current directory or in $HOME/.idefix.
func ReadConfig(name string) (*ClientOptions, error) {
	c := &ClientOptions{
		Meta: make(map[string]interface{}),
	}

	c.vp = viper.New()
	c.vp.SetConfigName(name)

	c.vp.AddConfigPath(".")
	c.vp.AddConfigPath("$HOME/.idefix/")

	ucd, err := os.UserConfigDir()
	if err != nil {
		ucd = "$HOME"
	}
	c.vp.AddConfigPath(ucd + "/idefix/")

	c.vp.SetDefault("Broker", "ssl://mqtt.terathings.com:8883")
	c.vp.SetDefault("Encoding", "mg")

	if err := c.vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			c.vp.SetConfigFile(ucd + "/idefix/" + name + ".json")
		}
		return c, err
	}

	if err := c.vp.Unmarshal(c); err != nil {
		return c, err
	}

	info, ok := debug.ReadBuildInfo()
	if ok {
		c.Meta["idefix-go"] = info.Main.Version
	}

	return c, nil
}

// UpdateConfig updates the configuration of the specified client. This function
// should only be called after the [ReadConfig] function has been executed successfully.
//
// Upon successful updating of the configuration, the changes are persisted by calling
// the WriteConfig method on the configuration provider.
func UpdateConfig(c *ClientOptions) error {
	if c.vp == nil {
		return fmt.Errorf("must ReadConfig first")
	}

	j, err := json.Marshal(c)
	if err != nil {
		return err
	}

	if err := c.vp.ReadConfig(bytes.NewReader(j)); err != nil {
		return err
	}

	return c.vp.WriteConfig()
}
