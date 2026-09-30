# Post-quantum TLS census

Scanned 10,000 sites on September 30, 2026. 7,340 answered over HTTPS; **4,086 of those (55.7%) negotiated a post-quantum key exchange.**

## By popularity

| Tranco rank | Sites | Post-quantum | Share |
|---|---:|---:|---:|
| Top 100 | 73 | 39 | 53.4% |
| Top 1,000 | 725 | 400 | 55.2% |
| Top 10,000 | 7,340 | 4,086 | 55.7% |

## By provider

The provider is whoever terminated TLS, detected from response headers.

| Provider | Sites | Post-quantum | Share |
|---|---:|---:|---:|
| Other | 3,283 | 671 | 20.4% |
| Cloudflare | 1,948 | 1,891 | 97.1% |
| Amazon CloudFront | 698 | 695 | 99.6% |
| Google | 362 | 208 | 57.5% |
| Akamai | 272 | 232 | 85.3% |
| Amazon (ELB/S3) | 272 | 12 | 4.4% |
| Fastly | 272 | 221 | 81.2% |
| Vercel | 74 | 71 | 95.9% |
| Unknown | 72 | 42 | 58.3% |
| Microsoft Azure | 47 | 32 | 68.1% |
| Netlify | 30 | 2 | 6.7% |
| GitHub | 10 | 9 | 90.0% |

## Key exchange groups

| Group | Sites | Post-quantum | Share |
|---|---:|---:|---:|
| X25519MLKEM768 | 4,086 | 4,086 | 100.0% |
| X25519 | 2,408 | 0 | 0.0% |
| CurveP256 | 720 | 0 | 0.0% |
| CurveP384 | 86 | 0 | 0.0% |
| CurveP521 | 40 | 0 | 0.0% |

## TLS versions

| Version | Sites | Post-quantum | Share |
|---|---:|---:|---:|
| TLS 1.3 | 6,205 | 4,086 | 65.9% |
| TLS 1.2 | 1,135 | 0 | 0.0% |

## Not reachable

| Reason | Sites |
|---|---:|
| dns | 1,411 |
| reset | 639 |
| tls | 251 |
| timeout | 231 |
| refused | 86 |
| blocked | 27 |
| other | 15 |

