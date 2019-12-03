package timectl

import (
	"sync"
	"time"
)

type TimeRingSeriesFactoryOption func(*TimeRingSeriesFactory)

func WithTimeRingSeriesFactoryTimeout(second int64) TimeRingSeriesFactoryOption {
	return func(t *TimeRingSeriesFactory) {
		t.timeout = second
	}
}

func WithTimeRingSeriesFactoryLength(second int) TimeRingSeriesFactoryOption {
	return func(t *TimeRingSeriesFactory) {
		t.length = second
	}
}

func NewTimeRingSeriesFactory(options ...TimeRingSeriesFactoryOption) *TimeRingSeriesFactory {
	t := &TimeRingSeriesFactory{
		ts:      make(map[interface{}]*TimeRingSeries),
		timeout: 5,
		length:  300,
	}
	for _, option := range options {
		option(t)
	}
	return t
}

type TimeRingSeriesFactory struct {
	rw      sync.RWMutex
	ts      map[interface{}]*TimeRingSeries
	timeout int64
	length  int
}

func (t *TimeRingSeriesFactory) LoadOrStore(key interface{}) *TimeRingSeries {
	t.rw.Lock()
	defer t.rw.Unlock()
	value := NewTimeRingSeries(WithTimeRingSecondLength(t.length))
	v, ok := t.ts[key]
	if ok {
		return v
	}
	t.ts[key] = value
	return value
}

func (t *TimeRingSeriesFactory) Last(key interface{}, ifempty Entry) Point {
	ts := t.LoadOrStore(key)
	now := time.Now().Unix()
	prev := ts.PrevN(2)
	if prev.Ctime+t.timeout < now {
		prev.Ctime = now
		prev.Entry = ifempty
	}
	return prev
}

func (t *TimeRingSeriesFactory) Points(key interface{}) []Point {
	var data = make([]Point, 0)
	now := time.Now().Unix()
	ts := t.LoadOrStore(key)
	for _, v := range ts.Points() {
		if v.Ctime > now-int64(ts.Len()) {
			data = append(data, v)
		}
	}
	return data
}

func (t *TimeRingSeriesFactory) Range(f func(key interface{}, v *TimeRingSeries) bool) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	for k, v := range t.ts {
		if !f(k, v) {
			break
		}
	}
}
