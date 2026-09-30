// Package scan connects to one site over HTTPS and records which TLS key
// exchange it negotiated.
//
// The client is Go's default TLS stack, which offers X25519MLKEM768 (hybrid
// post-quantum) first and plain X25519 alongside it, like current Chrome and
// Firefox. So the negotiated group is the server's choice: post-quantum if it
// supports it, classical if it doesn't.
package scan

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/netip"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/MasonKimball05/pq-census/internal/provider"
)

type Result struct {
	Rank       int    `json:"rank"`
	Domain     string `json:"domain"`
	Host       string `json:"host,omitempty"` // the host actually reached (the domain, or www.<domain>)
	OK         bool   `json:"ok"`
	Error      string `json:"error,omitempty"` // dns, timeout, refused, reset, tls, blocked, other
	TLSVersion string `json:"tls_version,omitempty"`
	Group      string `json:"group,omitempty"` // negotiated key exchange, e.g. X25519MLKEM768
	PQ         bool   `json:"pq"`
	Provider   string `json:"provider,omitempty"`
	Issuer     string `json:"issuer,omitempty"` // certificate issuer organization
	Ms         int64  `json:"ms,omitempty"`
}

// IsPostQuantum reports whether a negotiated group includes ML-KEM.
func IsPostQuantum(g tls.CurveID) bool {
	switch g {
	case tls.X25519MLKEM768, tls.SecP256r1MLKEM768, tls.SecP384r1MLKEM1024:
		return true
	}
	return strings.Contains(g.String(), "MLKEM")
}

type Scanner struct {
	client    *http.Client
	userAgent string
}

// New returns a Scanner. allowPrivate is for tests only: normally a domain
// that resolves to a private or loopback address is skipped, so a stray
// entry in the list can't make the scan poke at the local network.
func New(timeout time.Duration, userAgent string, allowPrivate bool) *Scanner {
	dialer := &net.Dialer{Timeout: timeout}
	if !allowPrivate {
		dialer.Control = func(_, address string, _ syscall.RawConn) error {
			ap, err := netip.ParseAddrPort(address)
			if err != nil {
				return err
			}
			a := ap.Addr().Unmap()
			if !a.IsGlobalUnicast() || a.IsPrivate() || netip.MustParsePrefix("100.64.0.0/10").Contains(a) {
				return errBlocked
			}
			return nil
		}
	}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		DisableKeepAlives:     true, // every site gets its own fresh handshake
		ForceAttemptHTTP2:     true,
	}
	return &Scanner{
		client: &http.Client{
			Transport: transport,
			Timeout:   2 * timeout,
			// Don't follow redirects: the handshake with the domain itself is
			// what's being measured, and the redirect's headers still show the
			// provider.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		userAgent: userAgent,
	}
}

var errBlocked = errors.New("resolves to a non-public address")

// Scan tries https://<domain>/, then https://www.<domain>/ if the bare domain
// can't be reached at all (many sites only serve on www).
func (s *Scanner) Scan(ctx context.Context, rank int, domain string) Result {
	r := s.try(ctx, rank, domain, domain)
	if !r.OK && (r.Error == "dns" || r.Error == "refused" || r.Error == "timeout") && !strings.HasPrefix(domain, "www.") {
		if w := s.try(ctx, rank, domain, "www."+domain); w.OK {
			return w
		}
	}
	return r
}

func (s *Scanner) try(ctx context.Context, rank int, domain, host string) Result {
	r := Result{Rank: rank, Domain: domain, Host: host}
	// The handshake is recorded as soon as it completes, so a site whose TLS
	// works but whose HTTP layer then fails (an HTTP/2 stream error, a HEAD it
	// refuses) still counts as a measurement.
	var state *tls.ConnectionState
	trace := &httptrace.ClientTrace{
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			if err == nil {
				state = &cs
			}
		},
	}
	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodHead, "https://"+host+"/", nil)
	if err != nil {
		r.Error = "other"
		return r
	}
	req.Header.Set("User-Agent", s.userAgent)

	start := time.Now()
	resp, err := s.client.Do(req)
	if err == nil {
		resp.Body.Close()
		if resp.TLS != nil {
			state = resp.TLS
		}
		r.Provider = provider.Detect(resp.Header)
	} else {
		r.Provider = "Unknown" // handshake done, but no response headers to go on
	}
	if state == nil {
		r.Provider = ""
		if err == nil {
			r.Error = "other"
		} else {
			r.Error = classify(err)
		}
		return r
	}

	r.OK = true
	r.Ms = time.Since(start).Milliseconds()
	r.TLSVersion = tls.VersionName(state.Version)
	r.Group = state.CurveID.String()
	r.PQ = IsPostQuantum(state.CurveID)
	if len(state.PeerCertificates) > 0 {
		if org := state.PeerCertificates[0].Issuer.Organization; len(org) > 0 {
			r.Issuer = org[0]
		}
	}
	return r
}

func classify(err error) string {
	var dnsErr *net.DNSError
	var certErr *tls.CertificateVerificationError
	var unknownAuth x509.UnknownAuthorityError
	var hostErr x509.HostnameError
	var alert tls.AlertError
	switch {
	case errors.Is(err, errBlocked):
		return "blocked"
	case errors.As(err, &dnsErr):
		return "dns"
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, os.ErrDeadlineExceeded):
		return "timeout"
	case isReset(err):
		// Reset mid-handshake. Usually a network filter keyed on the server
		// name (SNI), sometimes the server rejecting the client.
		return "reset"
	case errors.Is(err, syscall.ECONNREFUSED), errors.Is(err, syscall.EHOSTUNREACH), errors.Is(err, syscall.ENETUNREACH):
		return "refused"
	case errors.As(err, &certErr), errors.As(err, &unknownAuth), errors.As(err, &hostErr), errors.As(err, &alert):
		return "tls"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	if strings.Contains(err.Error(), "tls:") {
		return "tls"
	}
	return "other"
}

// wsaeconnreset is Windows' "connection forcibly closed by the remote host",
// which doesn't match syscall.ECONNRESET there.
const wsaeconnreset = 10054

func isReset(err error) bool {
	var errno syscall.Errno
	return errors.As(err, &errno) && (errno == syscall.ECONNRESET || errno == wsaeconnreset)
}

// String is a one-line summary for progress output.
func (r Result) String() string {
	if !r.OK {
		return fmt.Sprintf("%6d %-40s  error: %s", r.Rank, r.Domain, r.Error)
	}
	return fmt.Sprintf("%6d %-40s  %-16s %-8s %s", r.Rank, r.Domain, r.Group, r.TLSVersion, r.Provider)
}
