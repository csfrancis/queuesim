package main

import (
	"fmt"
	"sync"
	"time"
)

type Reactor struct {
	Clients []*Client
	Backend Backend
}

func (r *Reactor) Run() {
	wg := r.startClients()
	wg.Wait()

	for _, c := range r.Clients {
		fmt.Printf("%s\n", c)
	}
}

func (r *Reactor) startClients() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	numClients := len(r.Clients)
	clientCount := 0
	startCh := make(chan bool)

	for _, c := range r.Clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			clientCount += 1

			<-startCh

			c.Checkout(r.Backend)
		}(c)
	}

	for clientCount != numClients {
		time.Sleep(1 * time.Millisecond)
	}

	close(startCh)

	return wg
}
