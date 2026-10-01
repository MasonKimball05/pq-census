// Package probe is the follow-up study "can't, or won't?": for sites that
// chose a classical key exchange in the census, does the server lack
// post-quantum support, or does it support it and prefer classical?
//
// Each site gets two connections:
//
//	control: Go's default client (offers X25519MLKEM768 and X25519), as in
//	         the census. Re-checks the site is reachable and still classical.
//	pq-only: a client that offers only X25519MLKEM768. A server without
//	         post-quantum support shares no group with it and must refuse.
package probe

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/MasonKimball05/pq-census/internal/scan"
)

// Verdicts.
const (
	Supports     = "supports"     // pq-only handshake succeeded: supports it, prefers classical
	NoSupport    = "no-support"   // control fine, pq-only refused with a TLS alert: no PQ support
	NowPQ        = "now-pq"       // control negotiated PQ on every try: switched on since the census
	Mixed        = "mixed"        // control negotiated PQ on some tries but not others: a mixed server fleet
	Inconclusive = "inconclusive" // anything else: control failed, or pq-only failed without an alert
)

type Result struct {
	Rank        int           `json:"rank"`
	Domain      string        `json:"domain"`
	Host        string        `json:"host"`
	CensusGroup string        `json:"census_group"` // what it negotiated in the census
	Provider    string        `json:"provider"`     // as detected in the census
	Control     scan.Result   `json:"control"`
	PQOnly      scan.Result   `json:"pq_only"`
	Repeats     []scan.Result `json:"repeats,omitempty"` // extra control connections, when the first was PQ
	Verdict     string        `json:"verdict"`
	Reason      string        `json:"reason,omitempty"` // for inconclusive: why
}

// Candidates picks the census results worth probing: reached over TLS 1.3
// with a classical group. TLS 1.2 is left out because it has no way to
// negotiate a post-quantum key exchange at all.
func Candidates(census []scan.Result) []scan.Result {
	var out []scan.Result
	for _, r := range census {
		if r.OK && !r.PQ && r.TLSVersion == "TLS 1.3" && r.Group != "" {
			out = append(out, r)
		}
	}
	return out
}

// RepeatsForPQ is how many extra control connections a site gets when the
// first one negotiated post-quantum, to tell a site that switched it on from
// one whose servers disagree (a mixed fleet behind one name).
const RepeatsForPQ = 4

// noCommonGroupAlerts are the alerts RFC 8446 (section 4.1.1) requires when a
// server can't agree on parameters such as the key exchange group. Any other
// alert says something else went wrong, so it doesn't count as "no support".
var noCommonGroupAlerts = map[string]bool{"handshake failure": true, "insufficient security": true}

// Judge turns the connections into a verdict. repeats are the extra control
// connections made when the first control connection was post-quantum.
func Judge(control, pqOnly scan.Result, repeats []scan.Result) (verdict, reason string) {
	switch {
	case control.OK && control.PQ:
		for _, r := range repeats {
			if r.OK && !r.PQ {
				return Mixed, ""
			}
		}
		return NowPQ, ""
	case pqOnly.OK && pqOnly.PQ:
		return Supports, ""
	case !control.OK:
		return Inconclusive, "control failed: " + errorText(control)
	case pqOnly.Error == "tls" && noCommonGroupAlerts[pqOnly.Alert]:
		return NoSupport, ""
	case pqOnly.Error == "tls" && pqOnly.Alert != "":
		return Inconclusive, "pq-only refused with an unrelated alert: " + pqOnly.Alert
	default:
		return Inconclusive, "pq-only failed without a TLS alert: " + errorText(pqOnly)
	}
}

func errorText(r scan.Result) string {
	if r.Alert != "" {
		return r.Error + " (" + r.Alert + ")"
	}
	return r.Error
}

// ---- summary ----

type Count struct {
	Name         string  `json:"name"`
	Conclusive   int     `json:"conclusive"` // supports + no-support
	Supports     int     `json:"supports"`
	NoSupport    int     `json:"no_support"`
	SupportShare float64 `json:"support_share"` // supports / conclusive, in percent
}

type Summary struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Probed      int            `json:"probed"`
	Verdicts    map[string]int `json:"verdicts"`
	Overall     Count          `json:"overall"`
	ByProvider  []Count        `json:"by_provider"`
	ByGroup     []Count        `json:"by_census_group"`
	Alerts      map[string]int `json:"no_support_alerts"`
	Reasons     map[string]int `json:"inconclusive_reasons"`
}

func Summarize(results []Result, now time.Time) Summary {
	s := Summary{
		GeneratedAt: now.UTC(), Probed: len(results),
		Verdicts: map[string]int{}, Alerts: map[string]int{}, Reasons: map[string]int{},
	}
	providers, groups := map[string]*Count{}, map[string]*Count{}
	add := func(m map[string]*Count, name, verdict string) {
		c := m[name]
		if c == nil {
			c = &Count{Name: name}
			m[name] = c
		}
		tally(c, verdict)
	}
	s.Overall.Name = "All"
	for _, r := range results {
		s.Verdicts[r.Verdict]++
		switch r.Verdict {
		case Supports, NoSupport:
			tally(&s.Overall, r.Verdict)
			add(providers, r.Provider, r.Verdict)
			add(groups, r.CensusGroup, r.Verdict)
			if r.Verdict == NoSupport {
				s.Alerts[r.PQOnly.Alert]++
			}
		case Inconclusive:
			s.Reasons[r.Reason]++
		}
	}
	finish(&s.Overall)
	s.ByProvider, s.ByGroup = sorted(providers), sorted(groups)
	return s
}

func tally(c *Count, verdict string) {
	c.Conclusive++
	if verdict == Supports {
		c.Supports++
	} else {
		c.NoSupport++
	}
}

func finish(c *Count) {
	if c.Conclusive > 0 {
		c.SupportShare = float64(c.Supports) * 100 / float64(c.Conclusive)
	}
}

func sorted(m map[string]*Count) []Count {
	out := make([]Count, 0, len(m))
	for _, c := range m {
		finish(c)
		out = append(out, *c)
	}
	slices.SortFunc(out, func(a, b Count) int {
		return cmp.Or(cmp.Compare(b.Conclusive, a.Conclusive), cmp.Compare(a.Name, b.Name))
	})
	return out
}

func countTable(w io.Writer, title string, rows []Count, minSites int) {
	fmt.Fprintf(w, "| %s | Conclusive | Supports PQ | No PQ support | Supports share |\n|---|---:|---:|---:|---:|\n", title)
	for _, c := range rows {
		if c.Conclusive < minSites {
			continue
		}
		fmt.Fprintf(w, "| %s | %d | %d | %d | %.1f%% |\n", c.Name, c.Conclusive, c.Supports, c.NoSupport, c.SupportShare)
	}
	fmt.Fprintln(w)
}

func mapTable(w io.Writer, title string, m map[string]int) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b string) int { return cmp.Or(cmp.Compare(m[b], m[a]), cmp.Compare(a, b)) })
	fmt.Fprintf(w, "| %s | Sites |\n|---|---:|\n", title)
	for _, k := range keys {
		fmt.Fprintf(w, "| %s | %d |\n", k, m[k])
	}
	fmt.Fprintln(w)
}

func (s Summary) Markdown(w io.Writer) {
	o := s.Overall
	fmt.Fprintf(w, "# Can't, or won't? Post-quantum support among classical sites\n\n")
	fmt.Fprintf(w, "Probed %d sites that negotiated a classical key exchange over TLS 1.3 in the census, on %s. ",
		s.Probed, s.GeneratedAt.Format("January 2, 2006"))
	fmt.Fprintf(w, "Of the %d with a conclusive result, **%d (%.1f%%) support post-quantum key exchange but chose classical**, "+
		"and %d (%.1f%%) don't support it.\n\n", o.Conclusive, o.Supports, o.SupportShare, o.NoSupport, 100-o.SupportShare)
	if n, m := s.Verdicts[NowPQ], s.Verdicts[Mixed]; n+m > 0 {
		fmt.Fprintf(w, "%d more sites negotiated post-quantum by default this time. On %d connections each, %d did so every time "+
			"(switched on since the census) and %d only sometimes: their servers disagree, a fleet partly upgraded.\n\n",
			n+m, 1+RepeatsForPQ, n, m)
	}

	fmt.Fprintf(w, "## Verdicts\n\n")
	mapTable(w, "Verdict", s.Verdicts)
	fmt.Fprintf(w, "## By provider\n\nProviders with at least 10 conclusive results.\n\n")
	countTable(w, "Provider", s.ByProvider, 10)
	fmt.Fprintf(w, "## By the classical group chosen in the census\n\n")
	countTable(w, "Census group", s.ByGroup, 1)
	if len(s.Alerts) > 0 {
		fmt.Fprintf(w, "## How servers without support refused\n\n")
		mapTable(w, "TLS alert", s.Alerts)
	}
	if len(s.Reasons) > 0 {
		fmt.Fprintf(w, "## Inconclusive\n\nLeft out of the percentages.\n\n")
		mapTable(w, "Reason", s.Reasons)
	}
}
