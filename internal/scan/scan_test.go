package scan

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// serve starts an HTTPS server that only accepts the given key exchange groups
// and returns a Scanner that trusts it, plus the server's host:port.
func serve(t *testing.T, groups []tls.CurveID) (*Scanner, string) {
	t.Helper()
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Server", "cloudflare")
	}))
	srv.TLS = &tls.Config{CurvePreferences: groups}
	srv.StartTLS()
	t.Cleanup(srv.Close)

	s := New(2*time.Second, "test", true)
	s.client.Transport.(*http.Transport).TLSClientConfig = srv.Client().Transport.(*http.Transport).TLSClientConfig
	return s, strings.TrimPrefix(srv.URL, "https://")
}

func TestScanPostQuantum(t *testing.T) {
	s, host := serve(t, []tls.CurveID{tls.X25519MLKEM768, tls.X25519})
	r := s.try(context.Background(), 1, "example.test", host)
	if !r.OK || !r.PQ || r.Group != "X25519MLKEM768" || r.TLSVersion != "TLS 1.3" || r.Provider != "Cloudflare" {
		t.Fatalf("got %+v", r)
	}
}

func TestScanClassical(t *testing.T) {
	s, host := serve(t, []tls.CurveID{tls.X25519})
	r := s.try(context.Background(), 1, "example.test", host)
	if !r.OK || r.PQ || r.Group != "X25519" {
		t.Fatalf("got %+v", r)
	}
}

func TestScanBlocksPrivateAddresses(t *testing.T) {
	_, host := serve(t, nil)
	s := New(2*time.Second, "test", false)
	r := s.try(context.Background(), 1, "example.test", host) // 127.0.0.1
	if r.OK || r.Error != "blocked" {
		t.Fatalf("got %+v, want blocked", r)
	}
}

func TestIsPostQuantum(t *testing.T) {
	for g, want := range map[tls.CurveID]bool{
		tls.X25519MLKEM768: true, tls.SecP256r1MLKEM768: true, tls.SecP384r1MLKEM1024: true,
		tls.X25519: false, tls.CurveP256: false,
	} {
		if IsPostQuantum(g) != want {
			t.Errorf("IsPostQuantum(%v) = %v", g, !want)
		}
	}
}

// A server whose TLS works but whose HTTP layer fails still counts: the
// handshake is the measurement.
func TestScanKeepsHandshakeWhenHTTPFails(t *testing.T) {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler) // drop the connection without a response
	}))
	srv.TLS = &tls.Config{CurvePreferences: []tls.CurveID{tls.X25519MLKEM768}}
	srv.StartTLS()
	defer srv.Close()

	s := New(2*time.Second, "test", true)
	s.client.Transport.(*http.Transport).TLSClientConfig = srv.Client().Transport.(*http.Transport).TLSClientConfig
	r := s.try(context.Background(), 1, "example.test", strings.TrimPrefix(srv.URL, "https://"))
	if !r.OK || !r.PQ || r.Provider != "Unknown" {
		t.Fatalf("got %+v", r)
	}
}
