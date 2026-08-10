package cfgo

import (
	"fmt"
	"sync"
	"testing"
)

// TestConcurrentGetSetServesLatest verifies that once all writers have
// finished, Get returns the value from the last completed Set. A concurrent
// reader must never be able to poison future reads with an older value.
//
// Regression test for the stale-cache race introduced by the non-atomic
// RLock->Lock upgrade in Get (follow-up to the concurrent-map-writes panic
// fixed in c077667).
func TestConcurrentGetSetServesLatest(t *testing.T) {
	t.Chdir(t.TempDir())

	const rounds = 200
	const setsPerRound = 100

	for r := 0; r < rounds; r++ {
		cfg := New()
		const key = "CONCURRENCY_TEST_KEY"
		cfg.Set(key, "v0")

		stop := make(chan struct{})
		var wg sync.WaitGroup
		for g := 0; g < 4; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stop:
						return
					default:
						cfg.Get(key)
					}
				}
			}()
		}

		var final string
		for i := 0; i < setsPerRound; i++ {
			final = fmt.Sprintf("v%d", i)
			cfg.Set(key, final)
		}
		close(stop)
		wg.Wait()

		// All goroutines have finished; the last write was Set(key, final).
		if got := cfg.Get(key); got != final {
			t.Fatalf("round %d: after final Set(%q), Get returned %v — stale value served after all writers finished", r, final, got)
		}
	}
}

// TestConcurrentGetReloadConsistency verifies that Get cannot resurrect a key
// that a completed Reload removed: once Reload returns, Has and Get must
// agree with the reloaded state.
func TestConcurrentGetReloadConsistency(t *testing.T) {
	t.Chdir(t.TempDir())

	const rounds = 200
	for r := 0; r < rounds; r++ {
		cfg := New()
		src := NewMockConfigSource("mock", map[string]any{"RELOADED_KEY": "present"})
		cfg.AddSource(src)
		if err := cfg.Reload(); err != nil {
			t.Fatalf("initial reload: %v", err)
		}

		stop := make(chan struct{})
		var wg sync.WaitGroup
		for g := 0; g < 4; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stop:
						return
					default:
						cfg.Get("RELOADED_KEY")
					}
				}
			}()
		}

		src.data = map[string]any{} // next Reload drops the key
		if err := cfg.Reload(); err != nil {
			t.Fatalf("second reload: %v", err)
		}

		close(stop)
		wg.Wait()

		if cfg.Has("RELOADED_KEY") {
			t.Fatalf("round %d: Has reports a key the completed Reload removed", r)
		}
		if got := cfg.Get("RELOADED_KEY"); got != nil {
			t.Fatalf("round %d: Reload removed the key but Get returned %v (resurrected by a stale cache write)", r, got)
		}
	}
}
