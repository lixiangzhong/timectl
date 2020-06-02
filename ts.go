package timectl

import (
	"container/ring"
	"sync"
	"time"
)

//WithSeries  指定TS里的Series长度,决定实时图的时长
func WithSeries(n int) TSOption {
	return func(t *TS) {
		t.series = ring.New(n)
	}
}

type TSOption func(*TS)

func NewTS(options ...TSOption) *TS {
	ts := &TS{
		max:       make(map[interface{}]Object),
		seg:       ring.New(3),
		series:    nil,
		timestamp: time.Now().Unix(),
	}
	for _, option := range options {
		option(ts)
	}
	return ts
}

//TS  记录2个时间序列,一长一短,长的用来画实时图,短的来用取实时点,并计算max
type TS struct {
	mutex     sync.RWMutex
	max       map[interface{}]Object
	seg       *ring.Ring //seg.Value=Node
	series    *ring.Ring //series.Value=Node
	timestamp int64
}

//Node Node是TS.ring上的Value实例,用来记录同一个时序下的不同指标,例:同一个IP里的多个节点
type Node struct {
	data      map[interface{}]Object
	timestamp int64
}

//Push 往TS放入一个点,同时计算MergeKey的值,时间变动时计算max
func (t *TS) Push(value Object) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	now := time.Now().Unix()
	if now != t.timestamp { //时间变动时,计算max
		t.calcSeries()
		t.calcMax(value.Key())
		t.calcMax(value.MergeKey())
		t.seg = t.seg.Next()
		t.seg.Value = nil
	}
	t.timestamp = now
	node := t.getNode(t.seg)
	node.timestamp = now
	node.data[value.Key()] = value
	var sum Object
	for k, v := range node.data {
		if k != value.MergeKey() {
			sum = v.Sum(sum)
		}
	}
	node.data[value.MergeKey()] = sum
	t.seg.Value = node
}

//calcSeries  计算series,例子:实时图
func (t *TS) calcSeries() {
	if t.series != nil {
		node := t.getNode(t.seg)
		t.series = t.series.Next()
		nodets := t.getNode(t.series)
		nodets.data = node.data
		nodets.timestamp = node.timestamp
		t.series.Value = nodets
	}
}

//calcMax 计算max
func (t *TS) calcMax(key interface{}) {
	node := t.getNode(t.seg)
	val := node.data[key]
	max := t.max[key]
	if max == nil || max.Less(val) {
		t.max[key] = val
	}
}

//getNode 取r.Value的Node实例,并初使化
func (t *TS) getNode(r *ring.Ring) Node {
	node, _ := r.Value.(Node)
	if node.data == nil {
		node.data = make(map[interface{}]Object)
	}
	return node
}

//PopMax 取出max并清空
func (t *TS) PopMax() []Object {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	data := make([]Object, 0)
	for _, v := range t.max {
		data = append(data, v)
	}
	t.max = make(map[interface{}]Object)
	return data
}

//Max 取出指定key的Max
func (t *TS) Max(key interface{}) Object {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.max[key]
}

//Series  取出series,例子用来画实时图
func (t *TS) Series(obj Object) []ObjectPoint {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	data := make([]ObjectPoint, 0)
	if t.series == nil {
		return data
	}
	now := time.Now().Unix()
	start := now - int64(t.series.Len())
	end := now
	for i := start; i < end; i++ {
		data = append(data, ObjectPoint{
			Object: obj,
			Time:   i,
		})
	}
	p := t.series
	for {
		node := t.getNode(p)
		if node.timestamp >= start && node.timestamp < end {
			v, ok := node.data[obj.Key()]
			if ok {
				idx := node.timestamp - start
				data[idx].Object = v
			}
		}
		p = p.Prev()
		if p == t.series {
			break
		}
	}
	return data
}

//Get  取出指定timestamp下的key指标,node.data[key]
func (t *TS) Get(timestamp int64, obj Object) Object {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	p := t.seg
	for {
		node := t.getNode(p)
		if node.timestamp == timestamp {
			v, ok := node.data[obj.Key()]
			if !ok {
				return obj
			}
			return v
		}
		p = p.Prev()
		if p == t.seg {
			break
		}
	}
	return obj
}

func (t *TS) GetPoint(timestamp int64, obj Object) ObjectPoint {
	v := t.Get(timestamp, obj)
	return ObjectPoint{
		Object: v,
		Time:   timestamp,
	}
}

func (t *TS) GetLastPoint(obj Object) ObjectPoint {
	now := time.Now().Unix()
	return t.GetPoint(now-1, obj)
}

//GetLast Get(now-1)
func (t *TS) GetLast(obj Object) Object {
	now := time.Now().Unix()
	return t.Get(now-1, obj)
}

//List  取出指定timestamp下的所有指标,range node.data
func (t *TS) List(timestamp int64) []Object {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	data := make([]Object, 0)
	p := t.seg
	for {
		node := t.getNode(p)
		if node.timestamp == timestamp {
			for _, v := range node.data {
				data = append(data, v)
			}
			break
		}
		p = p.Prev()
		if p == t.seg {
			break
		}
	}
	return data
}
