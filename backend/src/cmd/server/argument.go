package server

import (
	"flag"
)

type Arguments struct {
	ConfigPath string
}

func parseArguments() (*Arguments, error) {
	args := &Arguments{}
	flag.StringVar(&args.ConfigPath, "config", "config.yaml", "Path to configuration file")
	flag.Parse()
	return args, nil
}
