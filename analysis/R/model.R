# Stage 5: does popularity matter once the CDN is accounted for?

# fit_rank_model(sites) -> a fitted glm
#
# Logistic regression of pq on log10(rank) and provider, fit on reachable
# TLS 1.3 sites only. Providers are grouped as in provider_table(): fewer than
# 30 reachable sites, counted before the TLS 1.3 filter, merged into "Other".
# "Other" is the reference level. (Counting after the filter happens to give
# the same groups in this data.)
# The formula must be pq ~ log10(rank) + provider.
fit_rank_model <- function(sites) {
  stop("Stage 5: fit_rank_model() isn't written yet")
}

# odds_ratios(fit, conf = 0.95) -> data.frame with columns
#   term, odds_ratio, lower, upper
# one row per coefficient except the intercept. lower and upper are Wald
# intervals on the log-odds scale, exponentiated.
odds_ratios <- function(fit, conf = 0.95) {
  stop("Stage 5: odds_ratios() isn't written yet")
}
