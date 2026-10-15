# Stage 4: does post-quantum support depend on the CDN?

# provider_table(sites, min_sites = 30) -> a table (or integer matrix)
#
# Rows: providers. Providers with fewer than `min_sites` reachable sites are
# merged into the "Other" row. Columns, in this order: "classical", "pq".
# Cells: counts of sites. `sites` is reachable sites only.
provider_table <- function(sites, min_sites = 30) {
  stop("Stage 4: provider_table() isn't written yet")
}

# chi_square(tab) -> list(statistic = , df = , p_value = )
# Pearson's chi-square test of independence, computed from the table by hand
# (expected counts, the statistic, its degrees of freedom), with pchisq() for
# the p-value. No continuity correction. Don't call chisq.test() here; the
# tests use it to check you.
chi_square <- function(tab) {
  stop("Stage 4: chi_square() isn't written yet")
}

# cramers_v(tab) -> a single number from 0 to 1, how strong the association is.
cramers_v <- function(tab) {
  stop("Stage 4: cramers_v() isn't written yet")
}
