package timectl

import (
	"sync"
	"time"
)

const (
	Day  = time.Hour * 24
	Week = Day * 7
	Year = Day * 365
)

type RetryLimitOption func(*RetryLimit)

func WithMaxInterval(d time.Duration) RetryLimitOption {
	return func(r *RetryLimit) {
		r.maxinterval = d
	}
}

func NewRetryLimit(options ...RetryLimitOption) *RetryLimit {
	r := RetryLimit{
		m:           make(map[string]retryLimitEntry),
		maxinterval: Day,
	}
	for _, option := range options {
		option(&r)
	}
	return &r
}

type RetryLimit struct {
	rw          sync.RWMutex
	m           map[string]retryLimitEntry
	maxinterval time.Duration
}

type retryLimitEntry struct {
	RetryAt  time.Time
	Interval time.Duration
}

func (r *RetryLimit) Failed(key string) {
	r.rw.Lock()
	defer r.rw.Unlock()
	v, ok := r.m[key]
	if !ok {
		v = retryLimitEntry{RetryAt: time.Now().Add(time.Second), Interval: time.Second}
	} else {
		v.Interval = v.Interval * 2
		if v.Interval > r.maxinterval {
			v.Interval = r.maxinterval
		}
		v.RetryAt = time.Now().Add(v.Interval)
	}
	r.m[key] = v
}

func (r *RetryLimit) Delete(key string) {
	r.rw.Lock()
	defer r.rw.Unlock()
	delete(r.m, key)
}

func (r *RetryLimit) Allow(key string) bool {
	r.rw.RLock()
	defer r.rw.RUnlock()
	v, ok := r.m[key]
	if !ok {
		return true
	}
	return time.Now().After(v.RetryAt)
}

func (r *RetryLimit) Wait(key string) {
	r.rw.RLock()
	v, ok := r.m[key]
	r.rw.RUnlock()
	if !ok {
		return
	}
	t := time.NewTimer(time.Until(v.RetryAt))
	defer t.Stop()
	<-t.C
}
