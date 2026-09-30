# Run of 2026-09-30

- List: Tranco top 1M, list ID `V349N` (https://tranco-list.eu/list/V349N), top 10,000 domains
- Vantage point: one US residential connection, macOS, Go 1.27.1 default TLS client
- Duration: 5m22s, 32 concurrent connections, 8s timeouts
- The network filters some content: 639 domains were reset mid-handshake
  (`"error": "reset"`); a partial run from a second machine on the same
  network saw the same resets. They're excluded from the percentages.
- Cross-check: a partial run (1,001 domains) from a second machine agreed on
  723 of 728 domains both reached.

`results.jsonl` is one JSON object per domain. `REPORT.md` and `summary.json`
are generated from it with `go run . report`.
