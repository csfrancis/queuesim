package main

import (
	"time"
)

type NoopBackend struct {
}

func (b *NoopBackend) Checkout(c *Client) bool {
	return true
}

func (b *NoopBackend) Poll(c *Client) (bool, time.Duration) {
	return true, 0
}
