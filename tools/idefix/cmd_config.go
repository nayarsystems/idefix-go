package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	idf "gitlab.com/garagemakers/idefix-go"
	"gopkg.in/yaml.v3"
)

type ConfigCmd struct {
	Load  LoadConfigCmd  `cmd:"" help:"Load config parameters from $HOME/.config/idefix/<profile>" name:"load"`
	Write StoreConfigCmd `cmd:"" help:"Write config parameters to $HOME/.config/idefix/<profile>" name:"store"`
}

type StoreConfigCmd struct {
	Filename      string `name:"profile" help:"Config name" required:"" arg:""`
	Address       string `name:"address" help:"Address" required:""`
	Token         string `name:"token" help:"Token" required:""`
	BrokerAddress string `name:"broker" help:"Broker Address" default:"ssl://mqtt.terathings.com:8883"`
	Encoding      string `name:"encoding" help:"Encoding" default:"mg"`
}

func (storecmd *StoreConfigCmd) Run(kctx *Context) error {

	options := idf.ClientOptions{
		BrokerAddress: storecmd.BrokerAddress,
		Address:       storecmd.Address,
		Token:         storecmd.Token,
		Encoding:      storecmd.Encoding,
		Meta:          map[string]interface{}{"Holi": true},
	}

	d, err := yaml.Marshal(options)
	if err != nil {
		return err
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir+"/idefix/", 0755); err != nil {
		return err
	}

	configPath := fmt.Sprintf("%s/idefix/%s.yaml", configDir, storecmd.Filename)

	fmt.Printf("Writing file to %s\n", configPath)
	return os.WriteFile(configPath, d, 0644)
}

type LoadConfigCmd struct {
	Filename string `name:"profile" help:"Config name" required:"" arg:""`
}

func (loadcmd *LoadConfigCmd) Run(kctx *Context) error {

	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	configPath := fmt.Sprintf("%s/idefix/%s.yaml", configDir, loadcmd.Filename)

	b, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Println("Profile not found. Available profiles:")
			finfo, err := ioutil.ReadDir(configDir + "/idefix/")
			if err != nil {
				return err
			}

			for _, v := range finfo {
				fmt.Printf(" - %s\n", strings.TrimSuffix(v.Name(), filepath.Ext(v.Name())))
			}

			return nil
		}
		return err
	}

	options := idf.ClientOptions{
		Meta: make(map[string]interface{}),
	}

	if err := yaml.Unmarshal(b, &options); err != nil {
		return err
	}

	fmt.Printf("Printing file %s\n", configPath)
	fmt.Printf("%#v\n", options)

	return nil
}
