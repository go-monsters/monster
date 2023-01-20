package memory

import "time"

type Item struct {
	val         interface{}
	createdTime time.Time
	lifespan    time.Duration
}

func (mi *Item) isExpire() bool {
	// 0 means forever
	if mi.lifespan == 0 {
		return false
	}
	return time.Since(mi.createdTime) > mi.lifespan
}
