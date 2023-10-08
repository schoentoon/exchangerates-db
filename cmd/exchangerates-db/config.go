package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config structure of the config file
type Config struct {
	Debug    bool   `yaml:"debug"`
	HttpAddr string `yaml:"addr"`
	DB       struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
	} `yaml:"database"`

	Importers []struct {
		Driver string     `yaml:"driver"`
		When   WhenConfig `yaml:"when"`
	} `yaml:"importers"`
}

// ReadConfig reads a file into the config structure
func ReadConfig(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	out := &Config{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&out)
	if err != nil {
		return nil, err
	}

	return out, err
}
