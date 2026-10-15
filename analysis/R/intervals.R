# Stage 3: how precise is a percentage?

# wald_ci(x, n, conf = 0.95) -> c(lower = , upper = )
# The textbook interval p +/- z * sqrt(p(1 - p) / n), as proportions (0 to 1).
# Don't clamp it to [0, 1]; seeing where it breaks is the point.
wald_ci <- function(x, n, conf = 0.95) {
  stop("Stage 3: wald_ci() isn't written yet")
}

# wilson_ci(x, n, conf = 0.95) -> c(lower = , upper = )
# The Wilson score interval, as proportions. Written from the formula in the
# guide, not by calling prop.test() or binom.test().
wilson_ci <- function(x, n, conf = 0.95) {
  stop("Stage 3: wilson_ci() isn't written yet")
}

# add_cis(summary, conf = 0.95) -> the same data.frame with two more columns,
# lower and upper, the Wilson interval for pq / sites as percentages (0 to 100).
# `summary` is the output of summarize_pq().
add_cis <- function(summary, conf = 0.95) {
  stop("Stage 3: add_cis() isn't written yet")
}
