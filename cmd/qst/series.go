package main

import (
	"math"
	"time"

	"github.com/prometheus/common/model"
)

type Response struct {
	Data   model.Matrix `json:"data"`
	ErrMsg string       `json:"err_msg"`
	Timing Timing       `json:"timing"`
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

func TrimSamples(decimal float64, matrix model.Matrix) {
	trimmer := TrimFloat(decimal)
	for _, stream := range matrix {
		for i, _ := range stream.Values {
			sample := &stream.Values[i]
			sample.Value = model.SampleValue(trimmer(float64(sample.Value)))
		}
	}
}
