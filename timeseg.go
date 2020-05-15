package timectl

import (
	"container/ring"
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

func WithRingLength(n int) TimeSegOption {
	return func(t *TimeSeg) {
		t.r = ring.New(n)
	}
}

type TimeSegOption func(*TimeSeg)

func NewTimeSeg(options ...TimeSegOption) *TimeSeg {
	t := &TimeSeg{duration: 300, timeout: 3, r: ring.New(3)}
	for _, option := range options {
		option(t)
	}
	return t
}

//TimeSeg  记录duration秒时间段内的数个点与最大值的点
type TimeSeg struct {
	mutex    sync.RWMutex
	max      Point
	r        *ring.Ring
	duration int64 //统计最大值有效时长
	timeout  int64 //ring数据点的超时
}

func (t *TimeSeg) Push(v Entry) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	now := time.Now().Unix()
	p := Point{
		Entry: v,
		Ctime: now,
	}
	t.r.Value = p
	t.r = t.r.Next()
	if now-t.max.Ctime > t.duration { //max过期
		t.max = p
		return
	}
	if !v.Less(t.max.Entry) {
		t.max = p
	}
}

func (t *TimeSeg) PopMax() Point {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	now := time.Now().Unix()
	max := t.max
	t.max.Entry = nil
	if now-max.Ctime > t.duration {
		max.Entry = nil
	}
	return max
}

func (t *TimeSeg) Prev() Point {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	now := time.Now().Unix()
	p := t.prevPoint(1)
	if now-p.Ctime > t.timeout {
		p.Entry = nil
	}
	p.Ctime = now
	return p
}

func (t *TimeSeg) PrevN(n int) Point {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	now := time.Now().Unix()
	p := t.prevPoint(n)
	if now-p.Ctime > t.timeout {
		p.Entry = nil
	}
	p.Ctime = now
	return p
}

func (t *TimeSeg) prevPoint(n int) Point {
	var p Point
	var r = t.r
	for i := 0; i < n; i++ {
		r = r.Prev()
	}
	p, _ = r.Value.(Point)
	return p
}

func (t *TimeSeg) Max() Point {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	now := time.Now().Unix()
	max := t.max
	if now-max.Ctime > t.duration {
		max.Entry = nil
	}
	return max
}
