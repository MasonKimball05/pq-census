# Rscript run_tests.R      every stage
# Rscript run_tests.R 3    just stage 3
args <- commandArgs(trailingOnly = TRUE)
filter <- if (length(args) > 0) paste0("^", args[1], "-") else NULL
testthat::test_dir("tests/testthat", filter = filter, reporter = "summary", stop_on_failure = FALSE)
