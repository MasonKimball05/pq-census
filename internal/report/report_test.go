package report

import (
	"strings"
	"testing"
	"time"

	"github.com/MasonKimball05/pq-census/internal/scan"
)

func TestBuild(t *testing.T) {
	var rs []scan.Result
	for i := 1; i <= 250; i++ {
		r := scan.Result{Rank: i, Domain: "d", OK: true, TLSVersion: "TLS 1.3", Group: "X25519", Provider: "Other"}
		if i <= 50 { // half of the top 100 are post-quantum, none after
			r.PQ, r.Group, r.Provider = true, "X25519MLKEM768", "Cloudflare"
		}
		if i > 240 {
			r = scan.Result{Rank: i, Domain: "d", Error: "dns"}
		}
		rs = append(rs, r)
	}
	s := Build(rs, time.Unix(0, 0))
	if s.Scanned != 250 || s.Reachable != 240 || s.PQ != 50 || s.Errors["dns"] != 10 {
		t.Fatalf("counts: %+v", s)
	}
	if len(s.Tiers) != 2 || s.Tiers[0].Name != "Top 100" || s.Tiers[0].Percent != 50 || s.Tiers[1].Name != "Top 250" {
		t.Fatalf("tiers: %+v", s.Tiers)
	}
	if s.Providers[0].Name != "Other" || s.Providers[1].Percent != 100 {
		t.Fatalf("providers: %+v", s.Providers)
	}
	var b strings.Builder
	s.Markdown(&b)
	if !strings.Contains(b.String(), "**50 of those (20.8%)") {
		t.Fatalf("markdown:\n%s", b.String())
	}
}

func TestCommas(t *testing.T) {
	for n, want := range map[int]string{7: "7", 100: "100", 1000: "1,000", 1234567: "1,234,567"} {
		if got := commas(n); got != want {
			t.Errorf("commas(%d) = %q, want %q", n, got, want)
		}
	}
}
