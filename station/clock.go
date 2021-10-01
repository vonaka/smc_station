package station

import (
	"sync"
	"time"
)

type Clock struct {
	sync.RWMutex
	start time.Time
}

func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) Start() {
	c.Lock()
	defer c.Unlock()
	c.start = time.Now()
}

func (c *Clock) Reset() {
	c.Start()
}

func (c *Clock) Time() int {
	c.RLock()
	defer c.RUnlock()
	return int(time.Now().Sub(c.start).Seconds())
}
