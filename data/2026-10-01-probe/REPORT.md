# Can't, or won't? Post-quantum support among classical sites

Probed 2119 sites that negotiated a classical key exchange over TLS 1.3 in the census, on October 1, 2026. Of the 1773 with a conclusive result, **2 (0.1%) support post-quantum key exchange but chose classical**, and 1771 (99.9%) don't support it.

18 more sites negotiated post-quantum by default this time. On 5 connections each, 18 did so every time (switched on since the census) and 0 only sometimes: their servers disagree, a fleet partly upgraded.

## Verdicts

| Verdict | Sites |
|---|---:|
| no-support | 1771 |
| inconclusive | 328 |
| now-pq | 18 |
| supports | 2 |

## By provider

Providers with at least 10 conclusive results.

| Provider | Conclusive | Supports PQ | No PQ support | Supports share |
|---|---:|---:|---:|---:|
| Other | 1542 | 2 | 1540 | 0.1% |
| Google | 151 | 0 | 151 | 0.0% |
| Netlify | 28 | 0 | 28 | 0.0% |
| Akamai | 22 | 0 | 22 | 0.0% |
| Amazon (ELB/S3) | 11 | 0 | 11 | 0.0% |
| Unknown | 11 | 0 | 11 | 0.0% |

## By the classical group chosen in the census

| Census group | Conclusive | Supports PQ | No PQ support | Supports share |
|---|---:|---:|---:|---:|
| X25519 | 1684 | 2 | 1682 | 0.1% |
| CurveP256 | 56 | 0 | 56 | 0.0% |
| CurveP384 | 28 | 0 | 28 | 0.0% |
| CurveP521 | 5 | 0 | 5 | 0.0% |

## How servers without support refused

| TLS alert | Sites |
|---|---:|
| handshake failure | 1771 |

## Inconclusive

Left out of the percentages.

| Reason | Sites |
|---|---:|
| pq-only failed without a TLS alert: other | 261 |
| pq-only failed without a TLS alert: reset | 37 |
| pq-only refused with an unrelated alert: no application protocol | 17 |
| control failed: timeout | 7 |
| pq-only refused with an unrelated alert: unexpected message | 3 |
| pq-only failed without a TLS alert: timeout | 2 |
| pq-only refused with an unrelated alert: protocol version not supported | 1 |

