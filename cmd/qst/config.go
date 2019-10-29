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
	Backends    []Backend         `yaml:"backends" json:"backends"`
	Headers     map[string]string `yaml:"headers" json:"headers"`
	Queries     []string          `yaml:"queries" json:"queries"`
	Start       time.Time         `yaml:"start" json:"start"`
	End         time.Time         `yaml:"end" json:"end"`
	N           int               `yaml:"n_rounds" json:"n_rounds"`
	LogLevel    string            `yaml:"log_level"  json:"log_level" `
	Parallelism int               `yaml:"parallelism" json:"parallelism"`
	StepSize    time.Duration
}

func (cfg *Config) Validate() {
	if cfg.Parallelism < 1 {
		cfg.Parallelism = 5
	}

	if cfg.StepSize == time.Duration(0) {
		cfg.StepSize = time.Minute
	}
}

// LoadConfig read YAML-formatted config from filename into cfg.
func LoadConfig(filename string, cfg *Config) error {
	type durationWrapper struct {
		StepSize string `yaml:"step_size" json:"step_size"`
	}

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "Error reading config file")
	}

	err = yaml.Unmarshal(buf, cfg)
	if err != nil {
		return errors.Wrap(err, "Error parsing config file")
	}

	var dur durationWrapper
	err = yaml.Unmarshal(buf, &dur)
	if err != nil {
		return errors.Wrap(err, "Error parsing config file")
	}

	parsedDur, err := time.ParseDuration(dur.StepSize)
	if err != nil {
		return err
	}

	cfg.StepSize = parsedDur

	return nil
}
