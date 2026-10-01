// pq-census measures how many popular websites negotiate a post-quantum TLS
// key exchange.
//
//	pq-census scan   -list top-1m.csv -n 10000 -out results.jsonl
//	pq-census report -in results.jsonl -md REPORT.md -json summary.json
//	pq-census probe  -in results.jsonl -out probe.jsonl
//	pq-census probe-report -in probe.jsonl -md PROBE.md -json probe-summary.json
//
// probe is the follow-up study "can't, or won't?": it revisits the sites that
// chose a classical key exchange and asks whether they support post-quantum
// at all (see internal/probe).
//
// The list is a Tranco CSV ("rank,domain" per line, https://tranco-list.eu).
// Scans resume: domains already in the output file are skipped.
package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MasonKimball05/pq-census/internal/probe"
	"github.com/MasonKimball05/pq-census/internal/report"
	"github.com/MasonKimball05/pq-census/internal/scan"
)

const userAgent = "pq-census/0.1 (research: post-quantum TLS adoption; +https://github.com/MasonKimball05/pq-census)"

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "scan":
		err = runScan(os.Args[2:])
	case "report":
		err = runReport(os.Args[2:])
	case "probe":
		err = runProbe(os.Args[2:])
	case "probe-report":
		err = runProbeReport(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "pq-census:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: pq-census scan -list top-1m.csv [-n 10000] [-out results.jsonl]
       pq-census report [-in results.jsonl] [-md REPORT.md] [-json summary.json]
       pq-census probe -in results.jsonl [-out probe.jsonl]
       pq-census probe-report [-in probe.jsonl] [-md PROBE.md] [-json probe-summary.json]`)
	os.Exit(2)
}

type entry struct {
	rank   int
	domain string
}

func readList(path string, n int) ([]entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = 2
	var out []entry
	for len(out) < n {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rank, err := strconv.Atoi(rec[0])
		if err != nil {
			continue // header line or junk
		}
		out = append(out, entry{rank, strings.ToLower(strings.TrimSpace(rec[1]))})
	}
	return out, nil
}

func readResults(path string) ([]scan.Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []scan.Result
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var r scan.Result
		if json.Unmarshal(sc.Bytes(), &r) == nil && r.Domain != "" {
			out = append(out, r)
		}
	}
	return out, sc.Err()
}

func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	list := fs.String("list", "", "Tranco CSV (rank,domain)")
	n := fs.Int("n", 10000, "scan the top N domains")
	out := fs.String("out", "results.jsonl", "results file (appended to; existing domains are skipped)")
	workers := fs.Int("workers", 32, "concurrent connections")
	timeout := fs.Duration("timeout", 8*time.Second, "per-connection timeout")
	quiet := fs.Bool("q", false, "don't print each result")
	fs.Parse(args)
	if *list == "" {
		return errors.New("-list is required")
	}

	entries, err := readList(*list, *n)
	if err != nil {
		return err
	}
	done := map[string]bool{}
	if prev, err := readResults(*out); err == nil {
		for _, r := range prev {
			done[r.Domain] = true
		}
	}
	var todo []entry
	for _, e := range entries {
		if !done[e.domain] {
			todo = append(todo, e)
		}
	}
	fmt.Fprintf(os.Stderr, "%d domains, %d already scanned, %d to go\n", len(entries), len(entries)-len(todo), len(todo))

	f, err := os.OpenFile(*out, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	s := scan.New(*timeout, userAgent, false)
	jobs := make(chan entry)
	results := make(chan scan.Result)
	var wg sync.WaitGroup
	for range *workers {
		wg.Go(func() {
			for e := range jobs {
				results <- s.Scan(ctx, e.rank, e.domain)
			}
		})
	}
	go func() {
		defer close(jobs)
		for _, e := range todo {
			select {
			case jobs <- e:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() { wg.Wait(); close(results) }()

	enc := json.NewEncoder(f)
	count, pq, start := 0, 0, time.Now()
	for r := range results {
		// A result cut short by Ctrl-C isn't a real measurement; leave it for
		// the next (resumed) run.
		if ctx.Err() != nil {
			continue
		}
		if err := enc.Encode(r); err != nil {
			return err
		}
		count++
		if r.PQ {
			pq++
		}
		if !*quiet {
			fmt.Println(r)
		} else if count%500 == 0 {
			fmt.Fprintf(os.Stderr, "%d/%d scanned (%.0fs)\n", count, len(todo), time.Since(start).Seconds())
		}
	}
	fmt.Fprintf(os.Stderr, "done: %d scanned this run, %d post-quantum, %s\n", count, pq, time.Since(start).Round(time.Second))
	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "interrupted: run the same command again to resume")
	}
	return nil
}

func runReport(args []string) error {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	in := fs.String("in", "results.jsonl", "results file from scan")
	md := fs.String("md", "", "write the Markdown report here (default: stdout)")
	js := fs.String("json", "", "also write the summary as JSON here")
	fs.Parse(args)

	results, err := readResults(*in)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return fmt.Errorf("no results in %s", *in)
	}
	s := report.Build(results, time.Now())

	w := io.Writer(os.Stdout)
	if *md != "" {
		f, err := os.Create(*md)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	s.Markdown(w)

	if *js != "" {
		b, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(*js, append(b, '\n'), 0o644)
	}
	return nil
}

func runProbe(args []string) error {
	fs := flag.NewFlagSet("probe", flag.ExitOnError)
	in := fs.String("in", "results.jsonl", "census results from scan")
	out := fs.String("out", "probe.jsonl", "probe results (appended to; existing domains are skipped)")
	workers := fs.Int("workers", 24, "concurrent sites (each gets two connections, one after the other)")
	timeout := fs.Duration("timeout", 8*time.Second, "per-connection timeout")
	quiet := fs.Bool("q", false, "don't print each result")
	redo := fs.String("redo", "", "comma-separated verdicts to probe again (e.g. now-pq), replacing their saved results")
	fs.Parse(args)

	if *redo != "" {
		if err := dropVerdicts(*out, strings.Split(*redo, ",")); err != nil {
			return err
		}
	}

	census, err := readResults(*in)
	if err != nil {
		return err
	}
	candidates := probe.Candidates(census)

	done := map[string]bool{}
	if prev, err := readProbeResults(*out); err == nil {
		for _, r := range prev {
			done[r.Domain] = true
		}
	}
	var todo []scan.Result
	for _, c := range candidates {
		if !done[c.Domain] {
			todo = append(todo, c)
		}
	}
	fmt.Fprintf(os.Stderr, "%d classical TLS 1.3 sites in the census, %d already probed, %d to go\n",
		len(candidates), len(candidates)-len(todo), len(todo))

	f, err := os.OpenFile(*out, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Two scanners: the control uses Go's default groups, like the census;
	// the probe offers only the hybrid post-quantum group.
	control := scan.New(*timeout, userAgent, false)
	pqOnly := scan.New(*timeout, userAgent, false).OfferOnly(tls.X25519MLKEM768)

	jobs := make(chan scan.Result)
	results := make(chan probe.Result)
	var wg sync.WaitGroup
	for range *workers {
		wg.Go(func() {
			for c := range jobs {
				ctl := control.Probe(ctx, c.Rank, c.Domain, c.Host)
				pq := pqOnly.Probe(ctx, c.Rank, c.Domain, c.Host)
				// A site classical in the census but PQ now: switched on, or
				// a mixed fleet? Ask a few more times.
				var repeats []scan.Result
				if ctl.OK && ctl.PQ {
					for range probe.RepeatsForPQ {
						repeats = append(repeats, control.Probe(ctx, c.Rank, c.Domain, c.Host))
					}
				}
				verdict, reason := probe.Judge(ctl, pq, repeats)
				results <- probe.Result{
					Rank: c.Rank, Domain: c.Domain, Host: c.Host, CensusGroup: c.Group, Provider: c.Provider,
					Control: ctl, PQOnly: pq, Repeats: repeats, Verdict: verdict, Reason: reason,
				}
			}
		})
	}
	go func() {
		defer close(jobs)
		for _, c := range todo {
			select {
			case jobs <- c:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() { wg.Wait(); close(results) }()

	enc := json.NewEncoder(f)
	count, start := 0, time.Now()
	verdicts := map[string]int{}
	for r := range results {
		if ctx.Err() != nil {
			continue // cut short by Ctrl-C: leave it for the resumed run
		}
		if err := enc.Encode(r); err != nil {
			return err
		}
		count++
		verdicts[r.Verdict]++
		if !*quiet {
			fmt.Printf("%6d %-40s %-14s %s\n", r.Rank, r.Domain, r.Verdict, r.Reason)
		} else if count%250 == 0 {
			fmt.Fprintf(os.Stderr, "%d/%d probed (%.0fs)\n", count, len(todo), time.Since(start).Seconds())
		}
	}
	fmt.Fprintf(os.Stderr, "done: %d probed this run in %s: %v\n", count, time.Since(start).Round(time.Second), verdicts)
	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "interrupted: run the same command again to resume")
	}
	return nil
}

func readProbeResults(path string) ([]probe.Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []probe.Result
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		var r probe.Result
		if json.Unmarshal(sc.Bytes(), &r) == nil && r.Domain != "" {
			out = append(out, r)
		}
	}
	return out, sc.Err()
}

func runProbeReport(args []string) error {
	fs := flag.NewFlagSet("probe-report", flag.ExitOnError)
	in := fs.String("in", "probe.jsonl", "results file from probe")
	md := fs.String("md", "", "write the Markdown report here (default: stdout)")
	js := fs.String("json", "", "also write the summary as JSON here")
	fs.Parse(args)

	results, err := readProbeResults(*in)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return fmt.Errorf("no results in %s", *in)
	}
	// Judge again from the saved connections, so the report always reflects
	// the current rules (e.g. which alerts count as "no support").
	for i := range results {
		r := &results[i]
		r.Verdict, r.Reason = probe.Judge(r.Control, r.PQOnly, r.Repeats)
	}
	s := probe.Summarize(results, time.Now())

	w := io.Writer(os.Stdout)
	if *md != "" {
		f, err := os.Create(*md)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	s.Markdown(w)

	if *js != "" {
		b, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(*js, append(b, '\n'), 0o644)
	}
	return nil
}

// dropVerdicts rewrites a probe results file without the results whose saved
// verdict is in verdicts, so those domains get probed again.
func dropVerdicts(path string, verdicts []string) error {
	results, err := readProbeResults(path)
	if err != nil {
		return err
	}
	drop := map[string]bool{}
	for _, v := range verdicts {
		drop[strings.TrimSpace(v)] = true
	}
	var b strings.Builder
	enc := json.NewEncoder(&b)
	kept := 0
	for _, r := range results {
		if !drop[r.Verdict] {
			if err := enc.Encode(r); err != nil {
				return err
			}
			kept++
		}
	}
	fmt.Fprintf(os.Stderr, "redo: dropped %d saved results to probe again\n", len(results)-kept)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
