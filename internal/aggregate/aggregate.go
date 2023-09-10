package aggregate

import (
	"sync"
	"time"
)

const (
	defaultCount    = 1000
	defaultSize     = 1 * 1024 * 1024 // 1MB
	defaultInterval = "5m"
)

type Config struct {
	Count    int    `json:"count"`
	Size     int    `json:"size"`
	Interval string `json:"interval"`
}

type aggregate struct {
	maxCount int
	count    int

	maxSize int
	size    int

	maxInterval time.Duration
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

	if time.Since(a.now) > a.maxInterval {
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

	if cfg.Interval == "" {
		cfg.Interval = defaultInterval
	}

	dur, err := time.ParseDuration(cfg.Interval)
	if err != nil {
		return nil, err
	}

	return &Aggregate{
		maxCount:    cfg.Count,
		maxSize:     cfg.Size,
		maxInterval: dur,
		mu:          sync.Mutex{},
		aggs:        make(map[string]*aggregate),
	}, nil
}

type Aggregate struct {
	maxCount    int
	maxSize     int
	maxInterval time.Duration

	mu   sync.Mutex
	aggs map[string]*aggregate
}

func (m *Aggregate) Add(key string, data []byte) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	agg, ok := m.aggs[key]
	if !ok {
		agg = &aggregate{
			maxCount:    m.maxCount,
			maxSize:     m.maxSize,
			maxInterval: m.maxInterval,
		}

		agg.Reset()
		m.aggs[key] = agg
	}

	return agg.Add(data)
}

func (m *Aggregate) Count(key string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	agg, ok := m.aggs[key]
	if !ok {
		return 0
	}

	return agg.Count()
}

func (m *Aggregate) Get(key string) [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()

	agg, ok := m.aggs[key]
	if !ok {
		return nil
	}

	return agg.Get()
}

func (m *Aggregate) GetAll() map[string]*aggregate {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.aggs
}

func (m *Aggregate) Reset(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	agg, ok := m.aggs[key]
	if !ok {
		return
	}

	agg.Reset()
}

func (m *Aggregate) ResetAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, agg := range m.aggs {
		agg.Reset()
	}
}
