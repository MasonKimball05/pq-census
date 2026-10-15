# Stage 6: are post-quantum handshakes slower?

# latency_compare(sites, provider = NULL) -> list(
#   n_pq = , n_classical = , median_pq = , median_classical = , p_value = )
#
# Compares handshake times (ms) of post-quantum and classical sites with a
# Wilcoxon rank-sum test. With `provider` set, only that provider's sites.
# `sites` is reachable sites only.
latency_compare <- function(sites, provider = NULL) {
  stop("Stage 6: latency_compare() isn't written yet")
}
