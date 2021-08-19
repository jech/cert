package cert

import (
	"sync"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	_, err := generate()
	if err != nil {
		t.Errorf("generate: %v", err)
	}
}

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := generate()
		if err != nil {
			b.Errorf("generate: %v", err)
		}
	}
}

func TestGet(t *testing.T) {
	cert := New("/no/such/file", "/no/such/file")
	cert1, err := cert.Get()
	if err != nil {
		t.Errorf("Get: %v", err)
	}

	cert2, err := cert.Get()
	if err != nil {
		t.Errorf("Get: %v", err)
	}

	if cert1 != cert2 {
		t.Errorf("cert1 != cert2")
	}
}

func TestGetParallel(t *testing.T) {
	cert := New("/no/such/file", "/no/such/file")
	cert1, err := cert.Get()
	if err != nil {
		t.Errorf("getCertificate: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			for i := 0; i < 1000; i++ {
				cert2, err := cert.Get()
				if err != nil {
					t.Errorf("Get: %v", err)
				}
				if cert1 != cert2 {
					t.Errorf("cert1 != cert2")
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestStoreParallel(t *testing.T) {
	cert := New("/no/such/file", "/no/such/file")
	done := make(chan int, 1)

	go func() {
		for i := 0; i < 8; i++ {
			_, err := cert.store(time.Time{}, time.Time{})
			if err != nil {
				t.Errorf("store: %v", err)
			}
		}
		close(done)
	}()

outer:
	for {
		select {
		case <-done:
			break outer
		default:
			_, err := cert.Get()
			if err != nil {
				t.Errorf("Get: %v", err)
			}
		}
	}
}

func BenchmarkGet(b *testing.B) {
	cert := New("/no/such/file", "/no/such/file")
	_, err := cert.Get()
	if err != nil {
		b.Errorf("getCertificate: %v", err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := cert.Get()
		if err != nil {
			b.Errorf("Get: %v", err)
		}
	}
}
