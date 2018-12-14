package main

import (
	"errors"
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// DefaultConfigPath refers to location of configuration mount dir specified in deployment
const DefaultConfigPath = "/etc/webhook/config/config.yaml"

// Config represents all config we need to initialize the webhook server
type Config struct {
	Trace Trace `yaml:"trace"`
}

// Trace is the configuration for the trace context added to pods
type Trace struct {
	SampleRate float64 `yaml:"sampleRate"`
}

// ParseConfigFromPath reads YAML config into config struct
func ParseConfigFromPath(c *Config, path string) (bool, error) {

	configYaml, err := ioutil.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("could not read YAML configuration file: %v", err)
	}

	err = yaml.Unmarshal(configYaml, &c)
	if err != nil {
		return false, fmt.Errorf("could not umarshal YAML configuration file: %v", err)
	}

	return true, nil
}

// Validate accepts a WebhookServerConfig and returns whether the config was valid and an error if needed
func (cfg *Config) Validate() (bool, error) {
	if cfg.Trace.SampleRate < 0 || cfg.Trace.SampleRate > 1 {
		return false, errors.New("sampling rate must be between 0 and 1 inclusive")
	}

	return true, nil
}
