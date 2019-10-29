package main

import (
	"time"

	"math"

	"github.com/cortexproject/cortex/pkg/querier/queryrange"
)

type QueryResult struct {
	Req    queryrange.PrometheusRequest `json:"Request"`
	Data   []queryrange.SampleStream    `json:"data"`
	Timing Timing                       `json:"timing"`
}

type Timing struct {
	Start time.Time // when query was initiated
	End   time.Time // when query finished

}

func (timing Timing) Duration() time.Duration {
	return timing.End.Sub(timing.Start)
}

func TrimFloat(decimal float64) func(x float64) float64 {
	coeff := math.Pow(10, decimal)
	return func(x float64) float64 {
		return math.Round(x*coeff) / coeff
	}
}

func TrimSamples(decimal float64, streams []queryrange.SampleStream) {
	trimmer := TrimFloat(decimal)
	for _, stream := range streams {
		for i, _ := range stream.Samples {
			sample := &stream.Samples[i]
			sample.Value = trimmer(sample.Value)
		}
	}
}
