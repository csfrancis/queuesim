package main

import (
	"time"
)

type Backend interface {
	Checkout(c *Client) (complete bool)
	Poll(c *Client) (complete bool, pollAfter time.Duration)
}
