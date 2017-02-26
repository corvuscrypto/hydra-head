package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

var paths = []string{"./config.yml", "./config.yaml"}

type config struct {
	Security struct {
		SecretKey string
	}
	Master struct {
		MasterAddress string
		MasterPort    string
	}
}

var globalConfig *config

func init() {
	var file *os.File
	for _, path := range paths {
		if f, err := os.Open(path); err == nil {
			file = f
			break
		}
	}
	if file == nil {
		fmt.Println("No config file specified or found!")
		os.Exit(1)
	}
	configData, _ := ioutil.ReadAll(file)
	err := yaml.Unmarshal(configData, &globalConfig)
	if err != nil {
		fmt.Println("Unable to read config file!", err)
		os.Exit(1)
	}
}
