# pq-census in R

A statistical look at the census of 2026-09-30 (`../data/2026-09-30/results.jsonl`):
how sure we can be of the 55.7% figure, whether post-quantum support really
depends on the CDN in front of a site, whether popularity matters once the CDN
is accounted for, and whether post-quantum handshakes are slower.

The Go scanner counted. This asks what the counts can and can't support.

## Running it

```bash
Rscript run_tests.R      # every stage's checks
Rscript run_tests.R 3    # just stage 3
Rscript report.R         # the figures and STATS.md (stage 7)
```

Start with [GUIDE.md](GUIDE.md). It has the stages, the statistics behind
each one, and what to write.

| Stage | File | What it answers |
|---|---|---|
| 1 | `R/load.R` | Reading the scan into a data frame |
| 2 | `R/describe.R` | How many sites, by CDN and by popularity |
| 3 | `R/intervals.R` | How precise is each percentage? |
| 4 | `R/independence.R` | Does post-quantum support depend on the CDN? |
| 5 | `R/model.R` | Does popularity matter, CDN held constant? |
| 6 | `R/latency.R` | Are post-quantum handshakes slower? |
| 7 | `report.R` | Figures and a written summary |
