package cfgo

import "testing"

func benchConfig() *config {
	c := &config{
		data:      make(map[string]any),
		overrides: make(map[string]any),
	}
	c.data["BENCH_STRING"] = "value"
	c.data["BENCH_INT"] = "8080"
	return c
}

func BenchmarkGet(b *testing.B) {
	c := benchConfig()
	for b.Loop() {
		_ = c.Get("BENCH_STRING")
	}
}

func BenchmarkLookup(b *testing.B) {
	c := benchConfig()
	for b.Loop() {
		_, _ = c.Lookup("BENCH_STRING")
	}
}

func BenchmarkGetInt(b *testing.B) {
	c := benchConfig()
	for b.Loop() {
		_ = c.GetInt("BENCH_INT")
	}
}

func BenchmarkGetAsInt(b *testing.B) {
	c := benchConfig()
	for b.Loop() {
		_, _ = GetAs[int](c, "BENCH_INT")
	}
}
