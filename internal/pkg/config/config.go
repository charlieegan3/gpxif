package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	GPXSource GPXSource `yaml:"gpx_source"`
}

type GPXSource struct {
	URLTemplate string `yaml:"url_template"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
}

func Load(configFile string) (Config, error) {
	var cfg Config
	var err error

	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return cfg, nil
}
