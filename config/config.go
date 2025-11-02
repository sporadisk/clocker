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
	Calc            *CalcConfig     `yaml:"calculator"`
}

type ExporterConfig struct {
	Name   string            `yaml:"name"`
	Params map[string]string `yaml:"params"`
}

type OutputConfig struct {
	Name   string            `yaml:"name"`
	Params map[string]string `yaml:"params"`
}

type CalcConfig struct {
	CategoryParseMode string `yaml:"categoryParseMode"`
}

func Load(path string) (*Config, error) {

	usingCustomConfigPath := (path != "")
	var err error
	if !usingCustomConfigPath {
		path, err = lookForConfig(".clocker.yaml")
		if err != nil {
			return nil, fmt.Errorf("lookForConfig: %w", err)
		}
	}

	conf := Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !usingCustomConfigPath {
			// No config was found, but no config path was specified either
			fmt.Println("No config file found - Using defaults")
			return &conf, nil // return an empty config
		}
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	fmt.Printf("Found config at: %s\n", path)

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	return &conf, nil
}

func lookForConfig(filename string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("os.UserHomeDir: %w", err)
	}

	paths := []string{
		filename,                 // look in the current working directory first
		homeDir + "/" + filename, // then look in the user's home directory
	}

	for _, path := range paths {
		finfo, err := os.Stat(path)

		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("os.Stat: %w", err)
		}

		if err != nil && errors.Is(err, os.ErrNotExist) {
			continue
		}

		if finfo != nil && finfo.IsDir() {
			continue
		}

		return path, nil
	}

	return "", nil
}
