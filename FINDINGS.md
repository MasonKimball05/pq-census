# Findings, in plain language

*Mason Kimball, October 2026. The full write-up, with charts, is at
[masonkimball.dev/projects/pq-census](https://masonkimball.dev/projects/pq-census/).
Methods, raw data and caveats are in this repo: the census in
[data/2026-09-30](data/2026-09-30/README.md), the follow-up in
[data/2026-10-01-probe](data/2026-10-01-probe/README.md).*

## Is your website quantum-safe?

It mostly depends on who hosts it, and if your host hasn't upgraded, there's
no switch to flip.

Encrypted traffic can be recorded today and decrypted years from now, once
quantum computers are strong enough. The fix already exists: a "hybrid" key
exchange that adds ML-KEM, the post-quantum algorithm NIST standardized in
2024 (FIPS 203). Chrome and Firefox already use it whenever the server
supports it.

Given this, I wanted to know: how many servers actually do? I wrote a small
Go tool, pq-census, and checked the 10,000 most popular websites.

## What the census found

- **56%** of the sites that answered (4,086 of 7,340) negotiate a
  post-quantum key exchange.
- **Your CDN decides it, not how big you are.** Behind Cloudflare, 97% are
  post-quantum; behind CloudFront, 99.6%. Self-hosted sites: about 20%.
  Popularity barely matters: 53% of the top 100 against 56% of the top
  10,000.
- **Even within one company it varies:** Amazon's CDN is nearly all
  post-quantum, while sites on its load balancers and S3 are about 4%.
- **15% of sites still run TLS 1.2**, which can't do post-quantum at all.

## The follow-up: can't, or won't?

My hypothesis going in was that many of the classical sites already supported
post-quantum and were simply set to prefer the older option, since the major
TLS libraries have been adding it.

To test that, I reconnected to the 2,119 sites that chose classical
encryption over TLS 1.3, this time offering *only* the post-quantum option. A
server that supports it can still connect; one that doesn't has to refuse.

**Of the 1,773 that gave a clear answer, 99.9% refused**, and an independent
TLS implementation (OpenSSL) agreed on every one I spot-checked. Only 2 had it
and chose not to use it. My hypothesis was wrong, and not by a little. And
most of them, 88%, are self-hosted: the same sites the census found lagging.

## Why it matters

The fix isn't convincing site owners to flip a switch. Their servers need
upgrading, and that's a much harder ask: it can mean moving to newer TLS
libraries, web servers or load balancers, some tied to an operating system or
a vendor's hardware, and testing all of it before anything changes in
production. For many teams that's a project with real cost, not a setting, so
it may be a while before these sites catch up. Meanwhile, their traffic can
still be recorded today.

## Surprises, and limits

- My own network was silently resetting about 6% of connections (a content
  filter, confirmed from a second machine), and Windows Defender flagged the
  scanner as a trojan. Both are documented in the data folders.
- A single connection measures one server, not a whole site: a few sites gave
  different answers an hour apart, and two accept Go's post-quantum-only
  handshake but reject OpenSSL's.
- It's one vantage point at one moment, so treat the numbers as a snapshot,
  not a census of the whole web.

## What's next

Post-quantum cryptography is what I want to study in grad school, and this
raised my next question: what's actually blocking these upgrades, and what's
the cheapest path through them? If you run servers, I'd love to hear what's in
your way. Open an issue on this repo, or reach me through
[masonkimball.dev](https://masonkimball.dev/contact/).
