// Package report turns scan results into summary numbers and a Markdown write-up.
package report

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/MasonKimball05/pq-census/internal/scan"
)

type Count struct {
	Name    string  `json:"name"`
	Sites   int     `json:"sites"`
	PQ      int     `json:"pq"`
	Percent float64 `json:"percent"` // PQ share of Sites
}

type Summary struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Scanned     int            `json:"scanned"`
	Reachable   int            `json:"reachable"`
	PQ          int            `json:"pq"`
	Percent     float64        `json:"percent"`
	Tiers       []Count        `json:"tiers"`     // top 100, top 1,000, ...
	Providers   []Count        `json:"providers"` // most sites first
	Groups      []Count        `json:"groups"`    // Sites = sites that negotiated this group
	TLSVersions []Count        `json:"tls_versions"`
	Errors      map[string]int `json:"errors"`
}

func pct(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) * 100 / float64(d)
}

func tally(m map[string]*Count, name string, pq bool) {
	c := m[name]
	if c == nil {
		c = &Count{Name: name}
		m[name] = c
	}
	c.Sites++
	if pq {
		c.PQ++
	}
}

func sorted(m map[string]*Count) []Count {
	out := make([]Count, 0, len(m))
	for _, c := range m {
		c.Percent = pct(c.PQ, c.Sites)
		out = append(out, *c)
	}
	slices.SortFunc(out, func(a, b Count) int {
		return cmp.Or(cmp.Compare(b.Sites, a.Sites), cmp.Compare(a.Name, b.Name))
	})
	return out
}

// Build summarizes results. Percentages are over reachable sites only.
func Build(results []scan.Result, now time.Time) Summary {
	s := Summary{GeneratedAt: now.UTC(), Scanned: len(results), Errors: map[string]int{}}
	providers, groups, versions := map[string]*Count{}, map[string]*Count{}, map[string]*Count{}

	maxRank := 0
	for _, r := range results {
		maxRank = max(maxRank, r.Rank)
		if !r.OK {
			s.Errors[r.Error]++
			continue
		}
		s.Reachable++
		if r.PQ {
			s.PQ++
		}
		tally(providers, r.Provider, r.PQ)
		tally(groups, r.Group, r.PQ)
		tally(versions, r.TLSVersion, r.PQ)
	}
	s.Percent = pct(s.PQ, s.Reachable)
	s.Providers, s.Groups, s.TLSVersions = sorted(providers), sorted(groups), sorted(versions)

	for limit := 100; ; limit *= 10 {
		limit = min(limit, maxRank)
		c := Count{Name: fmt.Sprintf("Top %s", commas(limit))}
		for _, r := range results {
			if r.OK && r.Rank <= limit {
				c.Sites++
				if r.PQ {
					c.PQ++
				}
			}
		}
		c.Percent = pct(c.PQ, c.Sites)
		s.Tiers = append(s.Tiers, c)
		if limit >= maxRank {
			break
		}
	}
	return s
}

func commas(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func table(w io.Writer, title string, rows []Count, limit int) {
	fmt.Fprintf(w, "| %s | Sites | Post-quantum | Share |\n|---|---:|---:|---:|\n", title)
	for i, c := range rows {
		if limit > 0 && i == limit {
			break
		}
		fmt.Fprintf(w, "| %s | %s | %s | %.1f%% |\n", c.Name, commas(c.Sites), commas(c.PQ), c.Percent)
	}
	fmt.Fprintln(w)
}

// Markdown writes a readable report.
func (s Summary) Markdown(w io.Writer) {
	fmt.Fprintf(w, "# Post-quantum TLS census\n\n")
	fmt.Fprintf(w, "Scanned %s sites on %s. %s answered over HTTPS; **%s of those (%.1f%%) negotiated a post-quantum key exchange.**\n\n",
		commas(s.Scanned), s.GeneratedAt.Format("January 2, 2006"), commas(s.Reachable), commas(s.PQ), s.Percent)

	fmt.Fprintf(w, "## By popularity\n\n")
	table(w, "Tranco rank", s.Tiers, 0)
	fmt.Fprintf(w, "## By provider\n\nThe provider is whoever terminated TLS, detected from response headers.\n\n")
	table(w, "Provider", s.Providers, 15)
	fmt.Fprintf(w, "## Key exchange groups\n\n")
	table(w, "Group", s.Groups, 0)
	fmt.Fprintf(w, "## TLS versions\n\n")
	table(w, "Version", s.TLSVersions, 0)

	if len(s.Errors) > 0 {
		fmt.Fprintf(w, "## Not reachable\n\n| Reason | Sites |\n|---|---:|\n")
		keys := make([]string, 0, len(s.Errors))
		for k := range s.Errors {
			keys = append(keys, k)
		}
		slices.SortFunc(keys, func(a, b string) int { return cmp.Compare(s.Errors[b], s.Errors[a]) })
		for _, k := range keys {
			fmt.Fprintf(w, "| %s | %s |\n", k, commas(s.Errors[k]))
		}
		fmt.Fprintln(w)
	}
}
