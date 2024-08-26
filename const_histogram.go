package metrics

import (
	"sync"
)

type ConstHistogram interface {
	Clear()
	Count() int64
	Sum() int64
	Mean() float64
	Bucket(float64) int64
	Buckets() map[float64]int64
	Snapshot() ConstHistogram
	Update(int64)
}

// GetOrRegisterConstHistogram returns an existing ConstHistogram or constructs and
// registers a new StandardConstHistogram.
func GetOrRegisterConstHistogram(name string, r Registry, buckets ...float64) ConstHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() ConstHistogram { return NewConstHistogram(buckets...) }).(ConstHistogram)
}

func NewConstHistogramFromExists(count, sum int64, buckets map[float64]int64) ConstHistogram {
	return &StandardConstHistogram{count: count, sum: sum, buckets: buckets}
}

// NewConstHistogram constructs a new StandardHistogram from a Sample.
func NewConstHistogram(buckets ...float64) ConstHistogram {
	if UseNilMetrics {
		return &NilConstHistogram{}
	}
	m := make(map[float64]int64, len(buckets))
	for _, b := range buckets {
		m[b] = 0
	}
	return &StandardConstHistogram{buckets: m}
}

// NewRegisteredConstHistogram constructs and registers a new StandardConstHistogram
func NewRegisteredConstHistogram(name string, r Registry, buckets ...float64) ConstHistogram {
	c := NewConstHistogram(buckets...)
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

type StandardConstHistogram struct {
	count   int64
	sum     int64
	buckets map[float64]int64
	lock    sync.RWMutex
}

func (h *StandardConstHistogram) Update(v int64) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.count++
	h.sum += v
	for k, b := range h.buckets {
		if float64(v) <= k {
			h.buckets[k] = b + 1
		}
	}
}
func (h *StandardConstHistogram) Count() int64 {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.count
}
func (h *StandardConstHistogram) Sum() int64 {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.sum
}
func (h *StandardConstHistogram) Bucket(bucket float64) int64 {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.buckets[bucket]
}
func (h *StandardConstHistogram) Buckets() map[float64]int64 {
	h.lock.RLock()
	defer h.lock.RUnlock()
	// copy一下
	buckets := make(map[float64]int64, len(h.buckets))
	for k, v := range h.buckets {
		buckets[k] = v
	}
	return buckets
}
func (h *StandardConstHistogram) Clear() {
	h.lock.RLock()
	defer h.lock.RUnlock()
	h.count = 0
	h.sum = 0
	for k := range h.buckets {
		h.buckets[k] = 0
	}
}
func (h *StandardConstHistogram) Mean() float64 {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if h.count == 0 {
		return 0.0
	}
	return float64(h.sum) / float64(h.count)
}
func (h *StandardConstHistogram) Snapshot() ConstHistogram {
	h.lock.RLock()
	defer h.lock.RUnlock()
	buckets := make(map[float64]int64, len(h.buckets))
	for k, v := range h.buckets {
		buckets[k] = v
	}
	return &ConstHistogramSnapshot{
		count:   h.count,
		sum:     h.sum,
		buckets: buckets,
	}
}

type ConstHistogramSnapshot struct {
	count   int64
	sum     int64
	buckets map[float64]int64
}

func (h *ConstHistogramSnapshot) Update(v int64) {
	panic("update on a snapshot is not supported")
}
func (h *ConstHistogramSnapshot) Count() int64 {
	return h.count
}
func (h *ConstHistogramSnapshot) Sum() int64 {
	return h.sum
}
func (h *ConstHistogramSnapshot) Bucket(bucket float64) int64 {
	return h.buckets[bucket]
}
func (h *ConstHistogramSnapshot) Buckets() map[float64]int64 {
	return h.buckets
}
func (h *ConstHistogramSnapshot) Clear() {
	panic("clear on a snapshot is not supported")
}
func (h *ConstHistogramSnapshot) Mean() float64 {
	if h.count == 0 {
		return 0.0
	}
	return float64(h.sum) / float64(h.count)
}
func (h *ConstHistogramSnapshot) Snapshot() ConstHistogram {
	return h
}

type NilConstHistogram struct {
}

func (h *NilConstHistogram) Clear()                      {}
func (h *NilConstHistogram) Count() int64                { return 0 }
func (h *NilConstHistogram) Mean() float64               { return 0.0 }
func (h *NilConstHistogram) Snapshot() ConstHistogram    { return &NilConstHistogram{} }
func (h *NilConstHistogram) Sum() int64                  { return 0 }
func (h *NilConstHistogram) Update(v int64)              {}
func (h *NilConstHistogram) Bucket(bucket float64) int64 { return 0 }
func (h *NilConstHistogram) Buckets() map[float64]int64  { return nil }
