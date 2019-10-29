package main

import (
	"io/ioutil"
	"time"

	"net/http"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	log "github.com/sirupsen/logrus"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sigs.k8s.io/yaml"
)

type Config struct {
	Hosts []struct {
		Host       string `yaml:"host"`
		Identifier string `yaml:"identifier"`
	} `yaml:"hosts"`
	Headers       map[string]string `yaml:"headers"`
	Queries       []string          `yaml:"queries"`
	Start         time.Time         `yaml:"start"`
	End           time.Time         `yaml:"end"`
	StepSize      int64             `yaml: "step_size"`
	StepsPerQuery int64             `yaml:"steps_per_query"`
	N             int               `yaml:"n_rounds"`
	LogLevel      string            `yaml:"log_level"`
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

type Runner struct {
	APIs   []API
	Cfg    Config
	Logger *log.Logger
}

type API struct {
	Identifier string
	Logger     *log.Entry
	v1.API
}

func NewRunner(conf Config) (*Runner, error) {
	runner := &Runner{
		Cfg:    conf,
		Logger: log.New(),
	}

	lvl, err := log.ParseLevel(conf.LogLevel)
	if err != nil {
		return nil, err
	}
	runner.Logger.SetLevel(lvl)

	for _, host := range conf.Hosts {
		config := api.Config{
			Address: host.Host,
		}

		if len(conf.Headers) > 0 {
			config.RoundTripper = promhttp.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				for key, value := range conf.Headers {
					req.Header.Add(key, value)
				}
				return http.DefaultTransport.RoundTrip(req)
			})
		}

		c, err := api.NewClient(config)
		if err != nil {
			return nil, err
		}

		runner.APIs = append(runner.APIs, API{
			Identifier: host.Identifier,
			Logger:     runner.Logger.WithFields(log.Fields{"identifier": host.Identifier}),
			API:        v1.NewAPI(c),
		})
	}

	return runner, nil

}
