package timectl

import (
	"container/ring"
	"sync"
	"time"
)

func WithTimeoutSecond(sec int64) TimeSegRingOption {
	return func(t *TimeSegRing) {
		t.timeout = sec
	}
}

func WithDurationSecond(sec int64) TimeSegRingOption {
	return func(t *TimeSegRing) {
		t.duration = sec
	}
}

func WithRingLength(n int) TimeSegRingOption {
	return func(t *TimeSegRing) {
		t.r = ring.New(n)
	}
}

type TimeSegRingOption func(*TimeSegRing)

func NewTimeSegRing(options ...TimeSegRingOption) *TimeSegRing {
	t := &TimeSegRing{duration: 300, timeout: 3, clock: &Clock{}, r: ring.New(3)}
	for _, option := range options {
		option(t)
	}
	return t
}

//TimeSegRing  记录duration秒时间段内的数个点与最大值的点
type TimeSegRing struct {
	mutex    sync.RWMutex
	clock    *Clock
	max      Point
	r        *ring.Ring
	duration int64 //统计最大值有效时长
	timeout  int64 //ring数据点的超时
}

func (d *TimeSegRing) Push(v Entry) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	now, insec := d.clock.SameSecond()
	p := Point{
		Entry: v,
		Ctime: now,
	}
	if !insec {
		d.r = d.r.Next()
	}
	d.r.Value = p
	if now-d.max.Ctime > d.duration { //max过期
		d.max = p
		return
	}
	if !v.Less(d.max.Entry) {
		d.max = p
	}
}

func (d *TimeSegRing) PopMax() Point {
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

func (d *TimeSegRing) Prev() Point {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	now := time.Now().Unix()
	p := d.prevPoint(1)
	if now-p.Ctime > d.timeout {
		p.Entry = nil
	}
	p.Ctime = now
	return p
}

func (d *TimeSegRing) PrevN(n int) Point {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	now := time.Now().Unix()
	p := d.prevPoint(n)
	if now-p.Ctime > d.timeout {
		p.Entry = nil
	}
	p.Ctime = now
	return p
}

func (t *TimeSegRing) prevPoint(n int) Point {
	var p Point
	var r = t.r.Prev()
	for i := 1; i < n; i++ {
		r = r.Prev()
	}
	p, _ = r.Value.(Point)
	return p
}

func (d *TimeSegRing) Max() Point {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	now := time.Now().Unix()
	max := d.max
	if now-max.Ctime > d.duration {
		max.Entry = nil
	}
	return max
}
