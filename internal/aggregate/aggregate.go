package aggregate

import (
	"time"
)

const (
	defaultCount    = 1000
	defaultSize     = 1 * 1024 * 1024 // 1MB
	defaultDuration = "5m"
)

type Config struct {
	Count    int    `json:"count"`
	Size     int    `json:"size"`
	Duration string `json:"duration"`
}

type aggregate struct {
	maxCount int
	count    int

	maxSize int
	size    int

	maxDuration time.Duration
	now         time.Time

	items [][]byte
}

func (a *aggregate) Reset() {
	a.count = 0
	a.size = 0
	a.now = time.Now()

	a.items = a.items[:0]
}

func (a *aggregate) Add(data []byte) bool {
	newCount := a.count + 1
	if newCount > a.maxCount {
		return false
	}

	newSize := a.size + len(data)
	if newSize > a.maxSize {
		return false
	}

	if time.Since(a.now) > a.maxDuration {
		return false
	}

	a.now = time.Now()
	a.count = newCount
	a.size = newSize
	a.items = append(a.items, data)

	return true
}

func (a *aggregate) Get() [][]byte {
	return a.items
}

func (a *aggregate) Count() int {
	return a.count
}

func (a *aggregate) Size() int {
	return a.size
}

func New(cfg Config) (*Aggregate, error) {
	if cfg.Count <= 1 {
		cfg.Count = defaultCount
	}

	if cfg.Size <= 1 {
		cfg.Size = defaultSize
	}

	if cfg.Duration == "" {
		cfg.Duration = defaultDuration
	}

	dur, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		return nil, err
	}

	return &Aggregate{
		maxCount:    cfg.Count,
		maxSize:     cfg.Size,
		maxDuration: dur,
		aggs:        make(map[string]*aggregate),
	}, nil
}

type Aggregate struct {
	maxCount    int
	maxSize     int
	maxDuration time.Duration

	aggs map[string]*aggregate
}

func (m *Aggregate) Add(key string, data []byte) bool {
	agg, ok := m.aggs[key]
	if !ok {
		agg = &aggregate{
			maxCount:    m.maxCount,
			maxSize:     m.maxSize,
			maxDuration: m.maxDuration,
		}

		agg.Reset()
		m.aggs[key] = agg
	}

	return agg.Add(data)
}

func (m *Aggregate) Count(key string) int {
	agg, ok := m.aggs[key]
	if !ok {
		return 0
	}

	return agg.Count()
}

func (m *Aggregate) Get(key string) [][]byte {
	agg, ok := m.aggs[key]
	if !ok {
		return nil
	}

	return agg.Get()
}

func (m *Aggregate) GetAll() map[string]*aggregate {
	return m.aggs
}

func (m *Aggregate) Reset(key string) {
	agg, ok := m.aggs[key]
	if !ok {
		return
	}

	agg.Reset()
}

func (m *Aggregate) ResetAll() {
	for _, agg := range m.aggs {
		agg.Reset()
	}
}
