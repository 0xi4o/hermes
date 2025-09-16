package main

import (
	"flag"

	"github.com/0xi4o/hermes/config"
	"github.com/0xi4o/hermes/server"
)

func setup_flags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the hermes server")
	flag.IntVar(&config.Port, "port", 7379, "port for the hermes server")
	flag.Parse()
}

func main() {
	setup_flags()
	server.RunSyncTCPServer()
}
