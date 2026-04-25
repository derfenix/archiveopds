package inpx

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestZipVolume_releaseLease_clampsUnderflow(t *testing.T) {
	t.Parallel()
	v := &zipVolume{}
	v.releaseLease() // 0 -> underflow, clamp
	if n := atomic.LoadInt32(&v.refs); n != 0 {
		t.Fatalf("expected refs=0 after underflow, got %d", n)
	}
	v.addLease()
	v.releaseLease()
	if n := atomic.LoadInt32(&v.refs); n != 0 {
		t.Fatalf("expected refs=0 after balanced add/release, got %d", n)
	}
}

func TestZipVolume_addReleaseLease_cycle(t *testing.T) {
	t.Parallel()
	v := &zipVolume{}
	for i := 0; i < 20; i++ {
		v.addLease()
	}
	for i := 0; i < 20; i++ {
		v.releaseLease()
	}
	if n := atomic.LoadInt32(&v.refs); n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestZipVolume_addRelease_lease_concurrent(t *testing.T) {
	t.Parallel()
	const n = 64
	var wg sync.WaitGroup
	v := &zipVolume{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				v.addLease()
				v.releaseLease()
			}
		}()
	}
	wg.Wait()
	if r := atomic.LoadInt32(&v.refs); r != 0 {
		t.Fatalf("expected refs=0, got %d", r)
	}
}
