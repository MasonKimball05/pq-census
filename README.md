# pq-census

How much of the popular web negotiates a **post-quantum TLS key exchange**?

`pq-census` connects to the top sites on the [Tranco list](https://tranco-list.eu)
and records the key exchange each one picks, along with the TLS version, the
certificate issuer, and the CDN or platform that terminated TLS.

Go, standard library only.

## Why this matters

A key exchange recorded today can be decrypted later, once a large enough
quantum computer exists ("harvest now, decrypt later"). The fix is a hybrid key
exchange such as **X25519MLKEM768**, which pairs classic X25519 with ML-KEM, the
lattice-based KEM that NIST standardized in 2024 as FIPS 203. The session stays
secure unless *both* parts are broken.

Browsers already offer it. Whether you get it depends on the server, and so on
who runs the server. This measures that.

## How it measures

- The client is Go's default TLS stack, which offers `X25519MLKEM768` first
  and plain `X25519` alongside it, like current Chrome and Firefox. The
  negotiated group is the **server's choice**: post-quantum if it supports it,
  classical if not.
- One `HEAD https://<domain>/` per site, with redirects *not* followed: the
  measurement is the handshake with the domain itself. If the bare domain
  doesn't answer at all, it tries `www.<domain>`.
- The provider comes from response headers (`Server`, `Cf-Ray`,
  `X-Amz-Cf-Id`, `X-Served-By`, ...), so it's whoever terminated TLS. Sites that
  run their own edge (Amazon, Apple, Wikipedia, ...) show up as "Other".
- Domains that resolve to private or loopback addresses are skipped.
- The user agent names the project and links here.

### Limits

- A single vantage point (one US residential connection) and a single moment.
  Large sites can answer differently by region or by load balancer.
- Header-based provider detection misses sites that strip those headers.
  When the handshake succeeds but the HTTP request after it fails, the
  handshake still counts and the provider is recorded as "Unknown".
- Tranco ranks domains by DNS and traffic signals, not only websites, so many
  top entries are infrastructure with no HTTPS site (`akamaiedge.net`,
  `gtld-servers.net`, `windowsupdate.com`). They show up as unreachable, and
  percentages are over reachable sites only.
- **Content filters show up as resets.** About 6% of the top 10,000
  (mostly adult sites, plus a few video, proxy and SaaS domains) were reset
  mid-handshake from the network this was run on, from two different machines,
  with post-quantum and classical handshakes alike. That's a filter on the
  network, not the sites. They count as unreachable (`reset`), so they're
  left out of the percentages. Local security software (Malwarebytes Web
  Protection, for one) can do the same, and Windows Defender's machine-learning
  detection may quarantine the scanner itself, because it opens hundreds of
  connections a minute.

## Usage

```bash
curl -LO https://tranco-list.eu/top-1m.csv.zip && unzip top-1m.csv.zip
go run . scan -list top-1m.csv -n 10000 -out results.jsonl -q
go run . report -in results.jsonl -md REPORT.md -json summary.json
```

Scans resume: stop with Ctrl-C and run the same command again, and domains
already in `results.jsonl` are skipped.

## Layout

```
main.go                     scan and report subcommands
internal/scan               one site: TLS handshake, group, provider, issuer
internal/provider           CDN / platform detection from headers
internal/report             summary numbers and the Markdown report
```

## Test

```bash
go test ./...
```

The scan tests run local TLS servers restricted to post-quantum or classical
groups and check what gets recorded.
