package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/garden/server"
	"github.com/pivotal-cf-experimental/garden-dot-net/backend"
	"github.com/pivotal-golang/lager"
)

var containerGraceTime = flag.Duration(
	"containerGraceTime",
	0,
	"time after which to destroy idle containers",
)
var tupperwareURL = flag.String(
	"tupperwareURL",
	"http://127.0.0.1",
	"URL for the Tupperware container server",
)

func main() {
	defaultListNetwork := "unix"
	defaultListAddr := "/tmp/garden.sock"
	if os.Getenv("PORT") != "" {
		defaultListNetwork = "tcp"
		defaultListAddr = "0.0.0.0:" + os.Getenv("PORT")
	}
	var listenNetwork = flag.String(
		"listenNetwork",
		defaultListNetwork,
		"how to listen on the address (unix, tcp, etc.)",
	)
	var listenAddr = flag.String(
		"listenAddr",
		defaultListAddr,
		"address to listen on",
	)

	logger := cf_lager.New("garden-dotnet")

	netBackend := backend.DotNetBackend{
		TupperwareURL: *tupperwareURL,
	}

	gardenServer := server.New(*listenNetwork, *listenAddr, *containerGraceTime, netBackend, logger)
	err := gardenServer.Start()
	if err != nil {
		logger.Fatal("Server Failed to Start", err)
		os.Exit(1)
	}

	logger.Info("started", lager.Data{
		"network": *listenNetwork,
		"addr":    *listenAddr,
	})

	signals := make(chan os.Signal, 1)

	go func() {
		<-signals
		gardenServer.Stop()
		os.Exit(0)
	}()

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	select {}
}
