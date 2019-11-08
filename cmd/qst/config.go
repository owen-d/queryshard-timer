package main

import (
	"io/ioutil"
	"time"

	"github.com/pkg/errors"

	"sigs.k8s.io/yaml"
)

type Backend struct {
	Host       string `yaml:"host" json:"host"`
	Identifier string `yaml:"identifier" json:"identifier"`
}

type Config struct {
	Backends    []Backend         `yaml:"backends" json:"backends"`
	Headers     map[string]string `yaml:"headers" json:"headers"`
	Queries     []string          `yaml:"queries" json:"queries"`
	Start       time.Time         `yaml:"start" json:"start"`
	End         time.Time         `yaml:"end" json:"end"`
	Splits      int               `yaml:"splits" json:"splits"`
	LogLevel    string            `yaml:"log_level"  json:"log_level" `
	Parallelism int               `yaml:"parallelism" json:"parallelism"`
	Cycles      int               `yaml:"cycles" json:"cycles"`
	StepSize    time.Duration
}

func (cfg *Config) Validate() {
	if cfg.Parallelism < 1 {
		cfg.Parallelism = 5
	}

	if cfg.StepSize == time.Duration(0) {
		cfg.StepSize = time.Minute
	}

	if cfg.Splits == 0 {
		cfg.Splits = 1
	}

	if cfg.Cycles == 0 {
		cfg.Cycles = 1
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
