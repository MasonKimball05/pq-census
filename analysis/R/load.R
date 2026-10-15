# Stage 1: reading the scan.
#
# results.jsonl has one JSON object per line, one line per domain. Reachable
# sites have tls_version, group, provider, issuer and ms; unreachable ones have
# error instead. Every line has rank, domain, host, ok and pq.

# load_census(path) -> data.frame, one row per line of the file, sorted by rank.
#
# Columns and types:
#   rank         integer    1 to 10000
#   domain       character
#   ok           logical    TRUE if the site answered over HTTPS
#   pq           logical    TRUE if it negotiated X25519MLKEM768
#   provider     character  NA when not reachable
#   tls_version  character  NA when not reachable
#   ms           numeric    handshake time in milliseconds, NA when not reachable
#   error        character  NA when reachable
#
# Other columns (host, group, issuer) may be kept or dropped.
load_census <- function(path) {
  stop("Stage 1: load_census() isn't written yet")
}

# reachable(census) -> the rows where ok is TRUE.
reachable <- function(census) {
  stop("Stage 1: reachable() isn't written yet")
}
