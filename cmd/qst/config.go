package main

import (
	"io/ioutil"
	"time"

	"github.com/pkg/errors"

	"sigs.k8s.io/yaml"
)

type Backend struct {
	Host       string `yaml:"host"`
	Identifier string `yaml:"identifier"`
}

type Config struct {
	Backends    []Backend         `yaml:"backends"`
	Headers     map[string]string `yaml:"headers"`
	Queries     []string          `yaml:"queries"`
	Start       time.Time         `yaml:"start"`
	End         time.Time         `yaml:"end"`
	N           int               `yaml:"n_rounds"`
	StepSize    time.Duration     `yaml:"step_size"`
	LogLevel    string            `yaml:"log_level"`
	Parallelism int               `yaml:"parallelism"`
}

func (cfg *Config) Validate() {
	if cfg.Parallelism < 1 {
		cfg.Parallelism = 5
	}
}

// LoadConfig read YAML-formatted config from filename into cfg.
func LoadConfig(filename string, cfg *Config) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "Error reading config file")
	}

	err = yaml.UnmarshalStrict(buf, cfg)
	if err != nil {
		return errors.Wrap(err, "Error parsing config file")
	}

	return nil
}
