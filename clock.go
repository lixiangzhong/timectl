package timectl

import (
	"sync"
	"time"
)

type Clock struct {
	mutex sync.Mutex
	t     int64
}

//SameSecond  同一秒
func (c *Clock) SameSecond() (int64, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	now := time.Now().Unix()
	if now == c.t {
		return c.t, true
	}
	c.t = now
	return c.t, false
}
