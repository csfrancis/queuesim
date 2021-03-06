package main

import (
	"sync"
	"time"
)

type Limiter struct {
	lock        sync.Locker
	limit       uint
	currentTime int64
	count       uint
	lastCount   uint
}

func NewLimiter(limit uint) *Limiter {
	return &Limiter{
		lock:  &sync.Mutex{},
		limit: limit,
	}
}

func (l *Limiter) Limit() uint {
	return l.limit
}

func (l *Limiter) Count() uint {
	return l.count
}

func (l *Limiter) LastCount() uint {
	return l.lastCount
}

func (l *Limiter) Incoming() bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	if time.Now().Unix() == l.currentTime {
		if l.count >= l.limit {
			return false
		} else {
			l.count += 1
			return true
		}
	} else {
		l.lastCount = l.count
		l.count = 1
		l.currentTime = time.Now().Unix()
		return true
	}
}
