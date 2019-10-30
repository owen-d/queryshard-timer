package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
)

type Runner struct {
	APIs   []API
	Cfg    Config
	Logger *log.Logger
	Timing Timing

	queue []*Request
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

	for _, backend := range conf.Backends {
		config := api.Config{
			Address: backend.Host,
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
			Identifier: backend.Identifier,
			Logger:     runner.Logger.WithFields(log.Fields{"backend": backend.Identifier}),
			API:        v1.NewAPI(c),
		})
	}

	return runner, nil

}

type RunnerResult struct {
	Backends []Backend      `json:"backends"`
	Queries  []*QueryResult `json:"queries"`
}

type QueryResult struct {
	Query  string
	Rounds []*Round `json:"rounds"`
}

type Round struct {
	Range     v1.Range `json:"range"`
	Responses map[string]*Response
	sync.Mutex
}

/*
{
  queries: {
    qry: "abc",
    rounds: [
      {
        start, end, etc,
        responses: {
          "backend1": {},
          "backend2": {},
	},
      }
    ]
  }
}
*/

func (r *Runner) Run() (res RunnerResult, error error) {
	res.Backends = r.Cfg.Backends
	r.Timing.Start = time.Now()

	for _, query := range r.Cfg.Queries {
		qRes := &QueryResult{
			Query: query,
		}

		rounds, err := CalculateRounds(query, r.Cfg.Start, r.Cfg.End, r.Cfg.StepSize, r.Cfg.N)
		r.Logger.Debug(fmt.Sprintf(
			"Calculated %d round from %s to %s with a stepsize of %s",
			r.Cfg.N, r.Cfg.Start, r.Cfg.End, r.Cfg.StepSize,
		))

		if err != nil {
			return res, err
		}

		for _, round := range rounds {
			for _, api := range r.APIs {
				req := &Request{
					Api:        api,
					Query:      query,
					Identifier: api.Identifier,
					Round:      round,
				}
				r.queue = append(r.queue, req)
			}
			qRes.Rounds = append(qRes.Rounds, round)
		}
		res.Queries = append(res.Queries, qRes)
	}

	r.Timing.End = time.Now()

	r.Process()

	r.Logger.WithField("num_requests", len(r.queue)).Debug("runner finished")
	return res, nil
}

func (r *Runner) Process() {
	r.Logger.Debug("Processing")
	// Use a waitgroup for all the workers to finish
	var wg sync.WaitGroup
	wg.Add(r.Cfg.Parallelism)

	// Feed all requests to a bounded intermediate channel to limit parallelism.
	intermediate := make(chan *Request)

	go func() {
		for i, req := range r.queue {
			intermediate <- req
			r.Logger.Debug(fmt.Sprintf("enqueued query %d of %d", i+1, len(r.queue)))
		}
		close(intermediate)
		r.Logger.Debug("closed intermediate chan")
	}()

	for i := 0; i < r.Cfg.Parallelism; i++ {
		go func() {
			for req := range intermediate {
				r.Logger.WithFields(log.Fields{
					"id":    req.Identifier,
					"range": req.Round.Range,
					"query": req.Query,
				}).Debug("Executing query")
				req.Execute()
			}
			wg.Done()

		}()
	}

	wg.Wait()
}

type Request struct {
	Api               API
	Query, Identifier string
	Round             *Round
}

func (req *Request) SetResp(resp *Response) {
	req.Round.Lock()
	defer req.Round.Unlock()

	req.Round.Responses[req.Identifier] = resp
}

func (req *Request) Execute() {
	resp := &Response{
		Timing: Timing{
			Start: time.Now(),
		},
	}

	val, _, err := req.Api.QueryRange(context.Background(), req.Query, req.Round.Range)
	resp.Timing.End = time.Now()
	if err != nil {
		resp.ErrMsg = err.Error()
		req.SetResp(resp)
		return
	}

	resp.Data = val.(model.Matrix)
	req.SetResp(resp)
	return
}

func CalculateRounds(query string, start, end time.Time, stepSize time.Duration, n int) (rounds []*Round, err error) {
	totalDur := end.Sub(start)
	if totalDur < 0 {
		return nil, fmt.Errorf("end < start")
	}

	perQueryDur := totalDur / time.Duration(n)

	if perQueryDur < stepSize {
		return nil, fmt.Errorf("stepsize too large or n too large to create %d queries with stepsize %v", n, stepSize)
	}

	for t := start; t.Before(end); t = t.Add(perQueryDur) {
		rounds = append(rounds, &Round{
			Responses: make(map[string]*Response),
			Range: v1.Range{
				Start: t,
				End:   t.Add(perQueryDur),
				Step:  stepSize,
			},
		})
	}

	return rounds, nil
}
