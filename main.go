package main

import (
	"os"
	"os/signal"
)

var numClients = 500

func setupInterruptHandler(r *Reactor) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case <-c:
			r.Summary()
		}
		os.Exit(1)
	}()
}

func main() {
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = NewClient()
	}

	reactor := &Reactor{
		Clients: clients,
		Backend: NewBinBackend(10),
	}

	setupInterruptHandler(reactor)

	reactor.Run()

	reactor.Summary()
}
