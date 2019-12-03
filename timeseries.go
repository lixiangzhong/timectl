package timectl

import (
	"container/ring"
	"sync"
)

type TimeRingSeriesOption func(*TimeRingSeries)

func WithTimeRingSecondLength(second int) TimeRingSeriesOption {
	return func(t *TimeRingSeries) {
		t.r = ring.New(second)
	}
}

func NewTimeRingSeries(options ...TimeRingSeriesOption) *TimeRingSeries {
	t := new(TimeRingSeries)
	t.r = ring.New(300)
	t.clock = &Clock{}
	for _, option := range options {
		option(t)
	}
	return t
}

type TimeRingSeries struct {
	mutex sync.RWMutex
	clock *Clock
	r     *ring.Ring
}

func (t *TimeRingSeries) Push(v Entry) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	now, inseconds := t.clock.SameSecond()
	prev := t.prevPoint(1)
	if inseconds {
		t.r.Prev().Value = v.Sum(prev.Entry)
		return
	}
	t.r.Value = Point{Entry: v, Ctime: now}
	t.r = t.r.Next()
}

func (t *TimeRingSeries) prevPoint(n int) Point {
	var p Point
	var r = t.r.Prev()
	for i := 1; i < n; i++ {
		r = r.Prev()
	}
	p, _ = r.Value.(Point)
	return p
}

func (t *TimeRingSeries) Prev() Point {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.prevPoint(1)
}

func (t *TimeRingSeries) PrevN(n int) Point {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.prevPoint(n)
}

func (t *TimeRingSeries) Points() []Point {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	data := make([]Point, 0)
	t.r.Do(func(v interface{}) {
		p, ok := v.(Point)
		if ok {
			data = append(data, p)
		}
	})
	return data
}

func (t *TimeRingSeries) Len() int {
	return t.r.Len()
}
