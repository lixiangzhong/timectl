package timectl

import (
	"sync"
	"time"
)

func WithTimeoutSecond(sec int64) TimeSegOption {
	return func(t *TimeSeg) {
		t.timeout = sec
	}
}

func WithDurationSecond(sec int64) TimeSegOption {
	return func(t *TimeSeg) {
		t.duration = sec
	}
}

type TimeSegOption func(*TimeSeg)

func NewTimeSeg(options ...TimeSegOption) *TimeSeg {
	t := &TimeSeg{duration: 300, timeout: 3}
	for _, option := range options {
		option(t)
	}
	return t
}

//TimeSeg  记录duration秒时间段内的最后一个点与最大的点
type TimeSeg struct {
	mutex    sync.RWMutex
	max      Point
	last     Point
	duration int64
	timeout  int64
}

func (d *TimeSeg) Push(v Entry) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	now := time.Now().Unix()
	if now-d.max.Ctime > d.duration {
		d.max.Entry = nil
	}
	d.last.Entry = v
	d.last.Ctime = now
	if !v.Less(d.max.Entry) {
		d.max.Entry = v
		d.max.Ctime = now
	}
}

func (d *TimeSeg) PopMax() Point {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	now := time.Now().Unix()
	max := d.max
	d.max.Entry = nil
	if now-max.Ctime > d.duration {
		max.Entry = nil
	}
	return max
}

func (d *TimeSeg) Last() Point {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	now := time.Now().Unix()
	p := d.last
	if now-d.last.Ctime > d.timeout {
		p.Entry = nil
	}
	p.Ctime = now
	return p
}

func (d *TimeSeg) Max() Point {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	now := time.Now().Unix()
	max := d.max
	if now-max.Ctime > d.duration {
		max.Entry = nil
	}
	return max
}
