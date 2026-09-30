// pq-census measures how many popular websites negotiate a post-quantum TLS
// key exchange.
//
//	pq-census scan   -list top-1m.csv -n 10000 -out results.jsonl
//	pq-census report -in results.jsonl -md REPORT.md -json summary.json
//
// The list is a Tranco CSV ("rank,domain" per line, https://tranco-list.eu).
// Scans resume: domains already in the output file are skipped.
package main

import (
	"bufio"
	"context"
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
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "pq-census:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: pq-census scan -list top-1m.csv [-n 10000] [-out results.jsonl]\n       pq-census report [-in results.jsonl] [-md REPORT.md] [-json summary.json]")
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
