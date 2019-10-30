package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

var (
	configFile string
	outFile    string
)

func init() {
	flag.StringVar(&configFile, "f", "", "path to config file")
	flag.StringVar(&outFile, "o", "", "path to file to write results")
	flag.Parse()
	if configFile == "" {
		panic(`unset configFile. try "-f <file>"`)
	}
	if outFile == "" {
		panic(`unset outFile. try "-o <file>"`)
	}
}

func main() {
	var conf Config
	err := LoadConfig(configFile, &conf)
	if err != nil {
		panic(err)
	}

	out, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}

	conf.Validate()

	runner, err := NewRunner(conf)
	runner.Logger.Debug(fmt.Sprintf("configured with %+v", conf))

	if err != nil {
		panic(err)
	}

	res, err := runner.Run()
	if err != nil {
		panic(err)
	}

	json.NewEncoder(out).Encode(res)
}
