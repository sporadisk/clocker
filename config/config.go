package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaulltFullDay string          `yaml:"defaultFullDay"`
	Exporter        *ExporterConfig `yaml:"exporter"`
	Output          *OutputConfig   `yaml:"output"`
}

type ExporterConfig struct {
	Name   string            `yaml:"name"`
	Params map[string]string `yaml:"params"`
}

type OutputConfig struct {
	Name   string            `yaml:"name"`
	Params map[string]string `yaml:"params"`
}

func Load(path string) (*Config, error) {
	var useDefaultConf bool
	useDefaultConf = (path == "")

	if useDefaultConf {
		path = ".clocker.yaml"
	}

	conf := Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && useDefaultConf {
			// No config was found, but no config path was specified either
			return &conf, nil // return an empty config
		}
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	return &conf, nil
}
