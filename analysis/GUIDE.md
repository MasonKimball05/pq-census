# Guide

You write the R. Each stage has a file in `R/` with the functions to write and
exactly what they must return, and a test file that checks them. Get a stage's
tests passing, read the "what to notice" part, then move on.

The census counted things. This project asks how much those counts can be
trusted and what they actually show, which is what statistics is for.

## Setup

```bash
brew install r
Rscript -e 'install.packages(c("jsonlite", "dplyr", "ggplot2", "testthat"), repos = "https://cloud.r-project.org")'
```

Then, from this folder:

```bash
Rscript run_tests.R 1   # stage 1's checks
```

Every test fails at first with "isn't written yet". That's expected.

An editor helps. RStudio (or Positron, its newer sibling) lets you run a line at
a time and see the data frame, which is the fastest way to learn R. VS Code with
the R extension works too.

### R in five minutes, coming from Python

- `<-` assigns (`=` works too, but R code uses `<-`). Vectors are everywhere:
  `x <- c(1, 2, 3)` and `x * 2` doubles every element, no loop.
- Indexing starts at 1. `x[1]` is the first element. `df[rows, cols]`, with
  either left blank for "all": `df[df$ok, ]` is the reachable rows.
- `df$column` is a column. `TRUE`/`FALSE`, and `NA` is missing (not `NULL`).
  Most functions return `NA` if any input is `NA` unless you pass `na.rm = TRUE`.
- `function(x, y = 2) { ... }` defines a function; the last expression is the
  return value.
- `?median` opens the help page for anything. Use it constantly.
- `dplyr` (in the setup) gives you `filter()`, `mutate()`, `group_by()`,
  `summarise()` and `arrange()`, joined with `|>`. Base R can do everything
  here too; use whichever reads better to you.

---

## Stage 1: loading the data (`R/load.R`)

**Write** `load_census()` and `reachable()`.

`results.jsonl` isn't one JSON document but one per line ("JSON Lines"). Look up
`jsonlite::stream_in()`; it reads exactly this format and fills fields a line
doesn't have with `NA`, which is what you want for `error` and `provider`.

Two things the tests will catch:

- **Types.** JSON has no integer type, but jsonlite guesses one when every value
  is whole. `ms` is always a whole number in this file, so it will come in as an
  integer. The contract says numeric (double), because a time is a measurement,
  not a count. `as.numeric()`.
- **Order.** The scanner wrote results as they finished, not in rank order.
  `order()`.

**What to notice.** Run `table(census$error, useNA = "ifany")`. Almost a quarter
of the top 10,000 never answered over HTTPS, and most of those are DNS failures.
Many "domains" in a popularity list are infrastructure (CDN hostnames, ad
servers) with no website at all. Every percentage in this project is "of the
sites that answered", and the report has to say so.

---

## Stage 2: describing it (`R/describe.R`)

**Write** `summarize_pq()` and `rank_band()`.

`summarize_pq()` is a grouped count. In dplyr that's `group_by()` then
`summarise()` with `n()` and `sum(pq)`. In base R, `aggregate()` or `tapply()`.
Note that `sum()` of a logical vector counts the `TRUE`s.

`rank_band()` turns ranks into three bands. `cut()` does it, and its `breaks`
and `labels` arguments are the whole job. The bands don't overlap: "101-1,000"
means ranks 101 through 1,000.

**What to notice.** Look at the bands. The top 100 sites are 53% post-quantum,
the next 900 are 55%, the rest 56%. Is popularity related to post-quantum
support at all? With only 73 sites in the top band, you can't tell yet from the
percentages alone. Stage 3 says how much a percentage from 73 sites can move.

---

## Stage 3: how precise is a percentage? (`R/intervals.R`)

**Write** `wald_ci()`, `wilson_ci()` and `add_cis()`.

55.7% is an estimate. A different list, a different day, or a different
network path would give a slightly different number. A 95% confidence interval
is a range built by a method that, used over and over, captures the true value
95% of the time.

**Wald**, the one in most intro courses. With p̂ = x/n and z the normal
quantile for your confidence level (`qnorm(1 - (1 - conf) / 2)`, 1.96 for 95%):

    p̂ ± z · √( p̂(1 − p̂) / n )

**Wilson**, the one statisticians recommend:

    center    = ( p̂ + z²/(2n) ) / ( 1 + z²/n )
    halfwidth = z · √( p̂(1 − p̂)/n + z²/(4n²) ) / ( 1 + z²/n )

    interval  = center ± halfwidth

Return `c(lower = ..., upper = ...)` so the tests can find each end by name.

`add_cis()` applies `wilson_ci()` to every row of a summary. `mapply()` or a
loop works. Remember the summary's `percent` is 0 to 100 and the interval is a
proportion.

**What to notice.**

- For the headline, Wald and Wilson agree to three decimals: 54.5% to 56.8%.
  With 7,340 sites and p̂ near a half, the choice doesn't matter.
- For CloudFront, 695 of 698, Wald's upper end is above 100%, which is
  impossible. Wald assumes the sampling distribution is a symmetric bell, and
  near 0% or 100% it isn't: it's squashed against the wall. Wilson is built by
  inverting a test, which keeps it inside [0, 1] and makes it lopsided in the
  right direction. Several CDNs in this data are near 100%, so this matters
  here, not just in a textbook.
- The top-100 band is 39 of 73: Wilson gives roughly 42% to 64%. That range
  contains every other band's rate, so the bands don't show popularity
  mattering. Stage 5 asks properly.

**Be ready to explain.** What a 95% interval does and doesn't mean (it's not
"95% chance the true value is in this range"). Why Wald fails at the edges. Why
these intervals only cover *sampling* error, not the scan's other limits, like
measuring from one network in one country.

---

## Stage 4: does support depend on the CDN? (`R/independence.R`)

**Write** `provider_table()`, `chi_square()` and `cramers_v()`.

The claim in the census is "the CDN decides it". A chi-square test of
independence asks: if post-quantum support had nothing to do with the provider,
how surprising would this table be?

For a table with r rows and c columns, total N:

- **Expected count** for each cell, if the two were independent:
  row total × column total / N. (`outer()` builds the whole matrix at once.)
- **Statistic**: the sum over all cells of (observed − expected)² / expected.
- **Degrees of freedom**: (r − 1)(c − 1).
- **p-value**: the chance of a statistic at least this large if independence
  were true. `pchisq(statistic, df, lower.tail = FALSE)`.

The test is only trustworthy when every expected count is at least about 5. That's
why `provider_table()` merges providers with fewer than 30 sites into "Other".
Merging them into their own "small providers" row instead would leave one
expected count at 4.4, too small.

**Worked example** (also in the tests). 40 people in group a, 60 in group b:

|   | classical | pq |
|---|---|---|
| a | 10 | 30 |
| b | 20 | 40 |

Expected for a/classical is 40 × 30 / 100 = 12. The statistic comes to 0.794
with 1 degree of freedom, p ≈ 0.37: no evidence the groups differ.

A tiny p-value says an association exists, not how strong it is; with thousands
of sites even a trivial difference gets a tiny p. **Cramér's V** measures
strength, from 0 (none) to 1 (the row tells you the column exactly):

    V = √( χ² / ( N · (min(r, c) − 1) ) )

**What to notice.** The census table gives a statistic over 4,000 on 10
degrees of freedom, so the p-value prints as 0. R can't represent a number that
small, and "p < 0.001" is the honest way to write it. V is about 0.75, a very
strong association. "The CDN decides it" survives a formal test.

**Be ready to explain.** Why the expected-count rule exists (the chi-square
distribution is an approximation that breaks with small counts). Why a p-value
alone isn't a measure of effect size.

---

## Stage 5: does popularity matter, CDN held constant? (`R/model.R`)

**Write** `fit_rank_model()` and `odds_ratios()`.

Stage 2's bands can't separate popularity from CDN: if popular sites are more
often on Cloudflare, they'd look more post-quantum even if popularity itself
did nothing. That's **confounding**, and regression is the standard tool for it.

**Logistic regression** models the log-odds of a yes/no outcome as a straight
line in the predictors:

    log( p / (1 − p) ) = β₀ + β₁ · log10(rank) + β(provider)

`glm(formula, family = binomial, data = ...)`. The provider column must be a
factor with "Other" as the reference level (`relevel()` or `factor(levels =)`),
so every provider's coefficient compares it to "Other".

Two modeling choices, both on purpose:

- **TLS 1.3 only.** Every post-quantum site in the data is on TLS 1.3; TLS 1.2
  has no way to negotiate ML-KEM. Including TLS 1.2 sites would mix "can't, by
  protocol" with "doesn't, by choice". Check it yourself: `table(sites$tls_version, sites$pq)`.
- **log10(rank).** The difference between rank 1 and 10 is more like the
  difference between 1,000 and 10,000 than between 9,991 and 10,000. One unit of
  log10(rank) is "ten times further down the list".

**Odds ratios.** A coefficient is a change in log-odds; `exp()` of it is an
odds ratio. An odds ratio of 2 for Cloudflare means a Cloudflare site's odds of
post-quantum are twice an "Other" site's, at the same rank. Intervals: take
`confint.default(fit)` (Wald intervals for the coefficients), then `exp()`.

**What to notice.** Look at the log10(rank) row. If its interval includes 1,
popularity has no detectable effect once the CDN is known, and the CDN really
is the story. The provider odds ratios will be enormous for Cloudflare and
CloudFront, with wide intervals for providers with few classical sites, which
is stage 3's lesson again in a new form.

**Be ready to explain.** Odds versus probability. What "held constant" means
in a regression. Why a model with providers in it can tell you something the
band percentages couldn't.

---

## Stage 6: are post-quantum handshakes slower? (`R/latency.R`)

**Write** `latency_compare()`.

ML-KEM key shares are large: about 1.2 KB from the client, over a kilobyte
back. Does that make handshakes measurably slower?

Handshake times are skewed (most are fast, a few take seconds), so compare
**medians**, and test with the **Wilcoxon rank-sum test** (`wilcox.test()`),
which compares ranks rather than raw times and isn't thrown off by a few huge
values the way a t-test is.

**What to notice.** This is the stage with the trap.

- Across all sites, post-quantum handshakes have a *much lower* median, about
  205 ms against 676 ms. Post-quantum makes handshakes faster?
- Within Cloudflare alone, it's 170 ms against 215 ms, a much smaller gap.

The overall comparison mostly measures *who serves the site*. Post-quantum
sites are mostly on big CDNs with servers near you; classical sites are mostly
self-hosted, farther away. That's confounding again, and it's why the
within-provider comparison is the fairer one. Even there, think about why the
gap might exist before calling it an effect of ML-KEM: are Cloudflare's
classical sites configured differently?

One measurement per site, from one network, is also thin evidence about
milliseconds. Say so in the report.

**Be ready to explain.** Why the median and a rank test suit skewed data.
Simpson's paradox: a comparison that flips or shrinks once you split by a third
variable.

---

## Stage 7: the report (`report.R`)

No tests here. `Rscript report.R` should write:

1. `figures/providers.png`: post-quantum share by provider with Wilson
   intervals as error bars (`ggplot2`: `geom_point()` plus `geom_errorbar()`,
   or `geom_col()`), providers ordered by share. `ggsave()`.
2. `figures/latency.png`: handshake times for post-quantum and classical
   sites, overall and within Cloudflare (boxplots or density curves; a log
   scale on the time axis helps).
3. `STATS.md`: a short write-up, in your own words, of each finding with its
   numbers. The headline with its interval. The chi-square result and V. What
   the model says about popularity. The latency comparison and why the overall
   one misleads. A "Limits" section (one vantage point, sites that never
   answered, a single day).

Then link it from the pq-census README, and it's done.

## When you're finished

You'll have done real statistics on data you collected, in R: confidence
intervals and why the textbook one fails, a test of independence and an
effect size, a logistic regression, a nonparametric test, and two run-ins with
confounding. Those are the things to name on the application and to be ready
to talk through.
