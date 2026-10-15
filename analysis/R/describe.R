# Stage 2: describing the data.

# summarize_pq(sites, by = NULL) -> data.frame
#
# `sites` is reachable sites only (the caller filters).
#
# With by = NULL: one row with columns sites (count), pq (count), and percent
# (pq / sites * 100).
#
# With by = "provider" (or any other column name): one row per value of that
# column, the column itself first, then sites, pq and percent. Sorted by sites,
# largest first; ties broken alphabetically.
summarize_pq <- function(sites, by = NULL) {
  stop("Stage 2: summarize_pq() isn't written yet")
}

# rank_band(rank) -> factor, the same length as `rank`, with levels in this order:
#   "1-100", "101-1,000", "1,001-10,000"
rank_band <- function(rank) {
  stop("Stage 2: rank_band() isn't written yet")
}
