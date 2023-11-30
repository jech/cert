package cert

import (
	"crypto/tls"
	"io"
	"net/http"
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

func BenchmarkTLS(b *testing.B) {
	cert := New("/no/such/file", "/no/such/file")

	l, err := tls.Listen("tcp", ":8443", &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return cert.Get()
		},
	})
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()

	go func() {
		buf := make([]byte, 100)
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			defer c.Close()
			c.Read(buf)
		}
	}()

	conf := &tls.Config{InsecureSkipVerify: true}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		conn, err := tls.Dial("tcp", "127.0.0.1:8443", conf)
		if err != nil {
			b.Fatal(err)
		}
		conn.Close()
	}
}

func BenchmarkHTTP(b *testing.B) {
	cert := New("/no/such/file", "/no/such/file")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{'\n'})
	})

	s := http.Server{
		Addr: ":8443",
		TLSConfig: &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return cert.Get()
			},
		},
		Handler: mux,
	}

	go s.ListenAndServeTLS("", "")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest("GET", "https://localhost:8443", nil)
		if err != nil {
			b.Fatal(err)
		}
		req.Close = true
		rep, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, rep.Body)
		rep.Body.Close()
	}

	s.Close()
}
