package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"golang.fsrv.services/eventdb/internal/mode"
)

type Mode interface {
	Initialize()
	Run(sig <-chan os.Signal)
}

func main() {
	var runner Mode
	flag.Parse()
	switch flag.Arg(0) {
	case "server":
		runner = &mode.Server{}
	default:
		fmt.Println("mode not implemented")
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// make flags parsable inside mode specific code
	os.Args = flag.Args()

	runner.Initialize()
	runner.Run(sigChan)
}
