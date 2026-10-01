package probe

import (
	"strings"
	"testing"
	"time"

	"github.com/MasonKimball05/pq-census/internal/scan"
)

var (
	classicalOK = scan.Result{OK: true, TLSVersion: "TLS 1.3", Group: "X25519"}
	pqOK        = scan.Result{OK: true, PQ: true, TLSVersion: "TLS 1.3", Group: "X25519MLKEM768"}
	refused     = scan.Result{Error: "tls", Alert: "handshake failure"}
	reset       = scan.Result{Error: "reset"}
	timeout     = scan.Result{Error: "timeout"}
)

func TestCandidates(t *testing.T) {
	census := []scan.Result{
		{Domain: "a", OK: true, TLSVersion: "TLS 1.3", Group: "X25519"},                   // yes
		{Domain: "b", OK: true, TLSVersion: "TLS 1.3", Group: "CurveP256"},                // yes
		{Domain: "c", OK: true, PQ: true, TLSVersion: "TLS 1.3", Group: "X25519MLKEM768"}, // already PQ
		{Domain: "d", OK: true, TLSVersion: "TLS 1.2", Group: "CurveP256"},                // TLS 1.2 can't
		{Domain: "e", Error: "dns"}, // unreachable
	}
	var got []string
	for _, r := range Candidates(census) {
		got = append(got, r.Domain)
	}
	if strings.Join(got, ",") != "a,b" {
		t.Fatalf("candidates = %v, want [a b]", got)
	}
}

func TestJudge(t *testing.T) {
	cases := []struct {
		name            string
		control, pqOnly scan.Result
		want            string
	}{
		{"supports but prefers classical", classicalOK, pqOK, Supports},
		{"no support: refused with an alert", classicalOK, refused, NoSupport},
		{"switched on since the census", pqOK, pqOK, NowPQ},
		{"control failed", timeout, refused, Inconclusive},
		{"pq-only reset, no alert", classicalOK, reset, Inconclusive},
		{"pq-only worked even though control didn't", timeout, pqOK, Supports},
	}
	for _, c := range cases {
		if got, _ := Judge(c.control, c.pqOnly, nil); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
	if _, reason := Judge(classicalOK, reset, nil); reason != "pq-only failed without a TLS alert: reset" {
		t.Errorf("reason = %q", reason)
	}

	// Only the alerts RFC 8446 requires for "no common parameters" mean no support.
	alpn := scan.Result{Error: "tls", Alert: "no application protocol"}
	if v, reason := Judge(classicalOK, alpn, nil); v != Inconclusive || reason != "pq-only refused with an unrelated alert: no application protocol" {
		t.Errorf("unrelated alert: got %s / %q", v, reason)
	}
	if v, _ := Judge(classicalOK, scan.Result{Error: "tls", Alert: "insufficient security"}, nil); v != NoSupport {
		t.Errorf("insufficient security: got %s", v)
	}

	// A PQ control is "now-pq" only if every repeat agrees; one classical repeat = mixed fleet.
	if v, _ := Judge(pqOK, refused, []scan.Result{pqOK, pqOK, pqOK, pqOK}); v != NowPQ {
		t.Errorf("consistent PQ: got %s", v)
	}
	if v, _ := Judge(pqOK, refused, []scan.Result{pqOK, classicalOK, pqOK, pqOK}); v != Mixed {
		t.Errorf("one classical repeat: got %s", v)
	}
}

func TestSummarize(t *testing.T) {
	mk := func(provider, group, verdict string, pq scan.Result) Result {
		return Result{Provider: provider, CensusGroup: group, Verdict: verdict, PQOnly: pq}
	}
	results := []Result{
		mk("Akamai", "X25519", Supports, pqOK),
		mk("Akamai", "X25519", Supports, pqOK),
		mk("Akamai", "X25519", NoSupport, refused),
		mk("Other", "CurveP256", NoSupport, refused),
		mk("Other", "X25519", NowPQ, pqOK),
		{Verdict: Inconclusive, Reason: "control failed: timeout"},
	}
	s := Summarize(results, time.Unix(0, 0))
	if s.Probed != 6 || s.Overall.Conclusive != 4 || s.Overall.Supports != 2 || s.Overall.SupportShare != 50 {
		t.Fatalf("overall: %+v", s.Overall)
	}
	if s.Verdicts[NowPQ] != 1 || s.Reasons["control failed: timeout"] != 1 || s.Alerts["handshake failure"] != 2 {
		t.Fatalf("verdicts %v reasons %v alerts %v", s.Verdicts, s.Reasons, s.Alerts)
	}
	if s.ByProvider[0].Name != "Akamai" || s.ByProvider[0].Supports != 2 || s.ByProvider[0].Conclusive != 3 {
		t.Fatalf("by provider: %+v", s.ByProvider)
	}

	var b strings.Builder
	s.Markdown(&b)
	if !strings.Contains(b.String(), "**2 (50.0%) support post-quantum key exchange but chose classical**") {
		t.Fatalf("markdown:\n%s", b.String())
	}
}
