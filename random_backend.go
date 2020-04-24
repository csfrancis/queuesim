package main

import (
	"sync/atomic"
	"time"
)

type RandomBackend struct {
	limiter               *Limiter
	checkoutCount         uint64
	throttleCheckoutCount uint64
}

func NewRandomBackend(rateLimit uint) *RandomBackend {
	return &RandomBackend{
		limiter: NewLimiter(rateLimit),
	}
}

func (b *RandomBackend) Checkout(c *Client) bool {
	if b.limiter.Incoming() {
		b.incrementCheckoutCount(c)
		return true
	}

	count := atomic.AddUint64(&b.throttleCheckoutCount, 1)
	c.SetData("checkout_throttled", count)
	return false
}

func (b *RandomBackend) Poll(c *Client) (bool, time.Duration) {
	if b.limiter.Incoming() {
		b.incrementCheckoutCount(c)
		return true, 0
	}
	return false, 1 * time.Second
}

func (b *RandomBackend) incrementCheckoutCount(c *Client) {
	count := atomic.AddUint64(&b.checkoutCount, 1)
	c.SetData("checkout_count", count)
}
