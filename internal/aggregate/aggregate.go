package aggregate

import (
	"time"
)

const (
	defaultCount    = 1000
	defaultSize     = 1 * 1024 * 1024 // 1MB
	defaultInterval = 5 * 60          // 5 minutes
)

type Config struct {
	Count    int `json:"count"`
	Size     int `json:"size"`
	Interval int `json:"interval"`
}

type Aggregate struct {
	maxCount    int
	maxSize     int
	maxInterval time.Duration

	count int
	size  int

	now   time.Time
	items [][]byte
}

func New(cfg Config) (*Aggregate, error) {
	if cfg.Count <= 1 {
		cfg.Count = defaultCount
	}

	if cfg.Size <= 1 {
		cfg.Size = defaultSize
	}

	if cfg.Interval <= 1 {
		cfg.Interval = defaultInterval
	}

	agg := &Aggregate{
		maxCount:    cfg.Count,
		maxSize:     cfg.Size,
		maxInterval: time.Duration(cfg.Interval) * time.Second,
	}

	agg.Reset()

	return agg, nil
}

func (a *Aggregate) Reset() {
	a.count = 0
	a.size = 0

	a.now = time.Now()
	a.items = a.items[:0]
}

func (a *Aggregate) Add(data []byte) bool {
	newCount := a.count + 1
	if newCount > a.maxCount {
		return false
	}

	newSize := a.size + len(data)
	if newSize > a.maxSize {
		return false
	}

	if time.Since(a.now) > a.maxInterval {
		return false
	}

	a.now = time.Now()
	a.count = newCount
	a.size = newSize
	a.items = append(a.items, data)

	return true
}

func (a *Aggregate) Get() [][]byte {
	return a.items
}

func (a *Aggregate) Count() int {
	return a.count
}

func (a *Aggregate) Size() int {
	return a.size
}
