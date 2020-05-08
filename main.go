package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

const defaultNumClients = 500
const defaultBackend = "bin"
const defaultRate = 10

func setupInterruptHandler(r *Reactor) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		var lastTime time.Time
		for {
			select {
			case <-c:
				now := time.Now()
				if now.Sub(lastTime) < 500*time.Millisecond {
					os.Exit(1)
				}
				lastTime = now

				fmt.Printf("\r")

				if backendSummary, ok := r.Backend.(BackendSummary); ok {
					backendSummary.Status()
				}
			}
		}
	}()
}

func main() {
	backend := flag.String("backend", defaultBackend, "queue backend: bin, noop, random")
	numClients := flag.Int("clients", defaultNumClients, "number of clients")
	rate := flag.Int("rate", defaultRate, "queue exit rate")

	flag.Parse()

	clients := make([]*Client, *numClients)
	for i := 0; i < *numClients; i++ {
		clients[i] = NewClient()
	}

	var backendImpl Backend
	switch *backend {
	case "bin":
		backendImpl = NewBinBackend(uint(*rate))
	case "random":
		backendImpl = NewRandomBackend(uint(*rate))
	case "noop":
		backendImpl = &NoopBackend{}
	}

	reactor := &Reactor{
		Clients: clients,
		Backend: backendImpl,
	}

	setupInterruptHandler(reactor)

	reactor.Run()

	reactor.Summary()
}
