package main

import (
	"fmt"
	"math"
	"sort"
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
}

func (r *Reactor) Summary() {
	for _, c := range r.Clients {
		fmt.Printf("%s\n", c)
	}

	r.showClientFairnessSummary()
}

func (r *Reactor) showClientFairnessSummary() {
	durations := make([]int, 0)
	for _, c := range r.Clients {
		d := c.data["considered_duration"]
		if d == nil {
			d, _ = time.ParseDuration("0s")
		}
		durations = append(durations, int(d.(time.Duration).Seconds()*1000))
	}
	sort.Ints(durations)

	p50 := durations[int(math.Floor(float64(len(durations))*0.50))]
	p75 := durations[int(math.Floor(float64(len(durations))*0.75))]
	p95 := durations[int(math.Floor(float64(len(durations))*0.95))]
	p99 := durations[int(math.Floor(float64(len(durations))*0.99))]

	fmt.Printf("considered_duration p50=%d,p75=%d,p95=%d,p99=%d\n", p50, p75, p95, p99)
}

func (r *Reactor) startClients() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	clientReady := &sync.WaitGroup{}
	startCh := make(chan bool)

	clientReady.Add(len(r.Clients))

	for _, c := range r.Clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			clientReady.Done()

			<-startCh

			c.Checkout(r.Backend)
		}(c)
	}

	clientReady.Wait()

	close(startCh)

	return wg
}
