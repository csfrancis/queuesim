package main

import (
	"fmt"
	"sync"
	"time"
)

const utilizationUpdateInterval = 500 * time.Millisecond
const binSize = 10
const initialWorkingBin = 2
const pollInterval = 1 * time.Second

type BinBackend struct {
	checkoutLimiter   *Limiter
	lock              sync.Locker
	utilization       float64
	binSize           int
	binLimiter        *Limiter
	queueBin          int
	queueBinPos       int
	workingBin        int
	workingBinUpdated time.Time
}

func NewBinBackend(rateLimit uint) *BinBackend {
	b := &BinBackend{
		checkoutLimiter: NewLimiter(rateLimit),
		lock:            &sync.Mutex{},
		binSize:         binSize,
		binLimiter:      NewLimiter(1 << 63),
		workingBin:      initialWorkingBin,
	}
	go b.utilizationProc()
	return b
}

func (b *BinBackend) Checkout(c *Client) bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.queueBinPos+1 > b.binSize {
		b.queueBin += 1
		b.queueBinPos = 0
	} else {
		b.queueBinPos += 1
	}

	c.SetData("bin", b.queueBin)

	return false
}

func (b *BinBackend) Poll(c *Client) (bool, time.Duration) {
	if c.GetData("bin").(int) > b.workingBin {
		return false, pollInterval
	}

	b.binLimiter.Incoming()

	if b.checkoutLimiter.Incoming() {
		if ct := c.GetData("considered_time"); ct != nil {
			c.SetData("considered_duration", time.Now().Sub(ct.(time.Time)))
		}
		return true, 0
	}

	if ct := c.GetData("considered_time"); ct == nil {
		c.SetData("considered_time", time.Now())
	}

	return false, pollInterval
}

func (b *BinBackend) calculateUtilization() {
	currentUtilization := float64(b.binLimiter.LastCount()) / float64(b.checkoutLimiter.Limit())
	if b.utilization == 0 {
		b.utilization = currentUtilization
	} else {
		b.utilization = currentUtilization*0.2 + b.utilization*0.8
	}
	fmt.Printf("utilization=%.3f\n", b.utilization)
	fmt.Printf("checkoutUtilization=%.3f\n", float64(b.checkoutLimiter.LastCount())/float64(b.checkoutLimiter.Limit()))
	fmt.Printf("workingBin=%d\n", b.workingBin)
}

func (b *BinBackend) shouldUpdateWorkingBin() bool {
	if b.utilization <= 2 {
		return true
	}

	return false
}

func (b *BinBackend) utilizationProc() {
	for {
		select {
		case <-time.After(utilizationUpdateInterval):
			b.calculateUtilization()

			if time.Now().Sub(b.workingBinUpdated) >= 1*time.Second {
				if b.shouldUpdateWorkingBin() {
					b.workingBin += 1
					b.workingBinUpdated = time.Now()
				}
			}
		}
	}
}
