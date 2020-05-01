package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

var clientId uint

type Client struct {
	id           uint
	startTime    time.Time
	endTime      time.Time
	pollRequests uint
	data         map[string]interface{}
}

func NewClient() *Client {
	clientId += 1
	return &Client{
		id:   clientId,
		data: make(map[string]interface{}),
	}
}

func (c *Client) Id() uint {
	return c.id
}

func (c *Client) StartTime() time.Time {
	return c.startTime
}

func (c *Client) EndTime() time.Time {
	return c.endTime
}

func (c *Client) CheckoutQueueDuration() time.Duration {
	return c.endTime.Sub(c.startTime)
}

func (c *Client) PollRequests() uint {
	return c.pollRequests
}

func (c *Client) GetData(key string) interface{} {
	return c.data[key]
}

func (c *Client) SetData(key string, value interface{}) {
	c.data[key] = value
}

func (c *Client) Checkout(backend Backend) {
	c.startTime = time.Now()

	completed := backend.Checkout(c)
	for !completed {
		c.pollRequests += 1

		completed, pollAfter := backend.Poll(c)
		if completed {
			break
		}
		<-time.After(pollAfter)
	}

	c.endTime = time.Now()
}

func makeTimestamp(t time.Time) int64 {
	return t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func (c *Client) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d", c.id, makeTimestamp(c.startTime), makeTimestamp(c.endTime),
		makeTimestamp(c.endTime)-makeTimestamp(c.startTime), c.pollRequests))

	keys := make([]string, 0)
	for k, _ := range c.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		str.WriteString(fmt.Sprintf(",%s=", k))
		if _, ok := c.data[k].(time.Time); ok {
			str.WriteString(fmt.Sprintf("%d", makeTimestamp(c.data[k].(time.Time))))
		} else {
			str.WriteString(fmt.Sprintf("%v", c.data[k]))
		}
	}
	return str.String()
}
