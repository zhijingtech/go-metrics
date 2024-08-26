package metrics

import "testing"

var buckets = []float64{1, 2, 3, 5, 7, 9}

func BenchmarkConstHistogram(b *testing.B) {
	h := NewConstHistogram(buckets...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Update(int64(i))
	}
}

func TestGetOrRegisterConstHistogram(t *testing.T) {
	r := NewRegistry()
	NewRegisteredConstHistogram("foo", r, buckets...).Update(47)
	if h := GetOrRegisterConstHistogram("foo", r, buckets...); 1 != h.Count() {
		t.Fatal(h)
	}
}

func TestConstHistogram10000(t *testing.T) {
	h := NewConstHistogram(buckets...)
	for i := 1; i <= 10000; i++ {
		h.Update(int64(i))
	}
	testConstHistogram10000(t, h)
}

func TestConstHistogramEmpty(t *testing.T) {
	h := NewConstHistogram(buckets...)
	if count := h.Count(); 0 != count {
		t.Errorf("h.Count(): 0 != %v\n", count)
	}
	if mean := h.Mean(); 0.0 != mean {
		t.Errorf("h.Mean(): 0.0 != %v\n", mean)
	}

	ps := h.Buckets()
	if 0.0 != ps[1] {
		t.Errorf("1: 0.0 != %v\n", ps[0])
	}
	if 0.0 != ps[3] {
		t.Errorf("3: 0.0 != %v\n", ps[1])
	}
	if 0.0 != ps[5] {
		t.Errorf("5: 0.0 != %v\n", ps[2])
	}
}

func TestConstHistogramSnapshot(t *testing.T) {
	h := NewConstHistogram(buckets...)
	for i := 1; i <= 10000; i++ {
		h.Update(int64(i))
	}
	snapshot := h.Snapshot()
	h.Update(0)
	testConstHistogram10000(t, snapshot)
}

func testConstHistogram10000(t *testing.T, h ConstHistogram) {
	if count := h.Count(); 10000 != count {
		t.Errorf("h.Count(): 10000 != %v\n", count)
	}

	if mean := h.Mean(); 5000.5 != mean {
		t.Errorf("h.Mean(): 5000.5 != %v\n", mean)
	}

	ps := h.Buckets()
	if 1 != ps[1] {
		t.Errorf("1: 1 != %v\n", ps[0])
	}
	if 3 != ps[3] {
		t.Errorf("3:3 != %v\n", ps[1])
	}
	if 5 != ps[5] {
		t.Errorf("5: 5 != %v\n", ps[2])
	}
}
