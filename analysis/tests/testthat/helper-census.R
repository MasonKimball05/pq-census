# Loads your code from analysis/R and gives the tests the census, read once.

analysis_dir <- normalizePath(file.path(testthat::test_path(), "..", ".."))
for (file in list.files(file.path(analysis_dir, "R"), pattern = "\\.R$", full.names = TRUE)) {
  source(file, local = TRUE)
}

census_path <- file.path(analysis_dir, "..", "data", "2026-09-30", "results.jsonl")

.cache <- new.env()
census <- function() {
  if (is.null(.cache$census)) .cache$census <- load_census(census_path)
  .cache$census
}
sites <- function() reachable(census())

# The worked example from the guide: two groups, 40 and 60 people.
example_table <- matrix(c(10, 20, 30, 40), nrow = 2,
                        dimnames = list(c("a", "b"), c("classical", "pq")))
