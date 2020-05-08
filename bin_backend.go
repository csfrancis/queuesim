package main

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

const utilizationUpdateInterval = 500 * time.Millisecond
const binSize = 10
const initialWorkingBin = 2
const pollInterval = 1 * time.Second
const minPollInterval = 1 * time.Second
const maxPollInterval = 30 * time.Second

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
	stats             []map[string]float64
}

func NewBinBackend(rateLimit uint) *BinBackend {
	b := &BinBackend{
		checkoutLimiter: NewLimiter(rateLimit),
		lock:            &sync.Mutex{},
		binSize:         binSize,
		binLimiter:      NewLimiter(1 << 63),
		workingBin:      initialWorkingBin,
		stats:           make([]map[string]float64, 0),
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

func calculatePollInterval(clientBin, workingBin int) time.Duration {
	pollInterval := minPollInterval
	diff := clientBin - workingBin

	if diff > 60 {
		pollInterval = 30
	} else if diff > 40 {
		pollInterval = 20
	} else if diff > 20 {
		pollInterval = 10
	} else if diff > 10 {
		pollInterval = 5
	} else if diff > 5 {
		pollInterval = 2
	}

	pollInterval *= time.Second

	if pollInterval > maxPollInterval {
		pollInterval = maxPollInterval
	}

	return pollInterval
}

func (b *BinBackend) Poll(c *Client) (bool, time.Duration) {
	clientBin := c.GetData("bin").(int)
	if clientBin > b.workingBin {
		return false, calculatePollInterval(clientBin, b.workingBin)
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

func (b *BinBackend) Status() {
	fmt.Printf("utilization=%.3f\n", b.utilization)
	fmt.Printf("checkoutUtilization=%.3f\n", float64(b.checkoutLimiter.LastCount())/float64(b.checkoutLimiter.Limit()))
	fmt.Printf("workingBin=%d/%d\n", b.workingBin, b.queueBin)
}

func (b *BinBackend) Summary() {
	cu := make([]float64, 0, len(b.stats))
	var sumCu float64
	for _, v := range b.stats {
		u := v["checkoutUtilization"]
		cu = append(cu, v["checkoutUtilization"])
		sumCu += u
	}
	sort.Float64s(cu)

	p05 := cu[int(math.Floor(float64(len(cu))*0.05))]
	p10 := cu[int(math.Floor(float64(len(cu))*0.10))]
	p20 := cu[int(math.Floor(float64(len(cu))*0.20))]

	fmt.Printf("checkoutUtilization avg=%.3f p05=%.3f p10=%.3f p20=%.3f\n",
		sumCu/float64(len(cu)), p05, p10, p20)
	fmt.Printf("final workingBin=%d\n", b.workingBin)
}

func (b *BinBackend) calculateUtilization() {
	currentUtilization := float64(b.binLimiter.LastCount()) / float64(b.checkoutLimiter.Limit())
	if b.utilization == 0 {
		b.utilization = currentUtilization
	} else {
		b.utilization = currentUtilization*0.2 + b.utilization*0.8
	}
}

func (b *BinBackend) shouldUpdateWorkingBin() bool {
	if b.utilization <= 2 {
		return true
	}

	return false
}

func (b *BinBackend) appendStats() {
	s := make(map[string]float64)
	s["checkoutUtilization"] = float64(b.checkoutLimiter.LastCount()) / float64(b.checkoutLimiter.Limit())
	b.stats = append(b.stats, s)
}

func (b *BinBackend) utilizationProc() {
	for {
		select {
		case <-time.After(utilizationUpdateInterval):
			b.calculateUtilization()

			if time.Now().Sub(b.workingBinUpdated) >= 1*time.Second {
				b.appendStats()

				if b.shouldUpdateWorkingBin() {
					b.workingBin += 1
					b.workingBinUpdated = time.Now()
				}
			}
		}
	}
}
