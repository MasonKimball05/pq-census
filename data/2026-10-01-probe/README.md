# Follow-up of 2026-10-01: can't, or won't?

The census of 2026-09-30 found 2,119 sites that negotiated a **classical**
key exchange over TLS 1.3. This study asks why: does the server lack
post-quantum support, or does it support it and prefer classical?

**Finding: it's almost always "can't".** Of 1,773 sites with a conclusive
result, **1,771 (99.9%) have no post-quantum support**, and only 2
(purdue.edu, roche.com) support it but chose classical.

## Method

Each site gets two connections, to the exact host the census reached:

| Connection | Client offers | Purpose |
|---|---|---|
| control | Go's defaults: X25519MLKEM768 and X25519 | the site is still reachable and still classical |
| pq-only | X25519MLKEM768 only | a server without post-quantum support shares no group, so it must refuse |

| Control | PQ-only | Verdict |
|---|---|---|
| classical | negotiates X25519MLKEM768 | **supports**: supports PQ but prefers classical |
| classical | refused with `handshake failure` or `insufficient security` | **no-support** |
| X25519MLKEM768 on all 5 tries | (any) | **now-pq**: switched on since the census |
| X25519MLKEM768 on some tries only | (any) | **mixed**: servers behind one name disagree |
| failed, or pq-only failed some other way | | **inconclusive**: left out of the percentages |

Only the two alerts **RFC 8446, section 4.1.1** requires for "no common
parameters" count as no support. A first pass also counted other alerts (for
example `no application protocol`, which is about HTTP version negotiation),
and OpenSSL showed those servers failing for an unrelated reason, so they
were moved to inconclusive.

## Results

| Verdict | Sites |
|---|---:|
| no-support | 1,771 |
| inconclusive | 328 |
| now-pq | 18 |
| supports | 2 |

Every no-support site refused with `handshake failure`. Full breakdowns by
provider and by census group are in [REPORT.md](REPORT.md).

## Cross-checks with a second TLS implementation

OpenSSL 3.6.5 (`s_client -groups X25519MLKEM768`) shares no code with Go's
TLS stack:

- **no-support:** a random 40 of 40 were refused by OpenSSL too.
- **supports:** both accepted OpenSSL's post-quantum-only handshake.
- **dropped connections** (inconclusive "other"): a random 30 of 30 failed
  under OpenSSL too, without a usable answer.
- **now-pq:** 16 of 18 accepted OpenSSL as well, on 3 tries each.
  **tn.com.ar and infobae.com accept Go's post-quantum-only handshake but
  reject OpenSSL's every time**, though both offer the same single group: the
  server treats the clients differently, probably a CDN or firewall
  fingerprinting the client. They count as now-pq, since they demonstrably
  support it.

## Caveats

- **Server fleets aren't uniform.** 4 sites negotiated post-quantum on the
  first pass, then answered classically (and refused post-quantum-only) on a
  second pass an hour later. One name can front servers at different upgrade
  levels, so a single connection measures one server, not the site.
- **The "now-pq" 18 may not all be new.** A site counted classical in the
  census could have had post-quantum servers all along, with the census
  reaching an older one.
- **One vantage point** (a US residential connection), as in the census.

## Reproduce

```bash
go run . probe -in data/2026-09-30/results.jsonl -out data/2026-10-01-probe/results.jsonl -q
go run . probe -in data/2026-09-30/results.jsonl -out data/2026-10-01-probe/results.jsonl -redo now-pq -q
go run . probe-report -in data/2026-10-01-probe/results.jsonl -md data/2026-10-01-probe/REPORT.md -json data/2026-10-01-probe/summary.json
```

The second command re-probes the sites that looked newly post-quantum, with
the repeat connections that tell "switched on" from "mixed fleet".
