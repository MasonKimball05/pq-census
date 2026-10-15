test_that("the provider table merges small providers into Other", {
  tab <- provider_table(sites())
  expect_equal(colnames(tab), c("classical", "pq"))
  expect_equal(sum(tab), 7340)
  expect_equal(sum(tab[, "pq"]), 4086)
  expect_false("GitHub" %in% rownames(tab))   # 10 sites, under 30
  expect_equal(nrow(tab), 11)
  expect_equal(unname(tab["Other", ]), c(2613, 680))
  expect_equal(unname(tab["Cloudflare", ]), c(57, 1891))
})

test_that("every expected count is at least 5, so the test is valid", {
  tab <- provider_table(sites())
  expected <- outer(rowSums(tab), colSums(tab)) / sum(tab)
  expect_gte(min(expected), 5)
})

test_that("chi_square() on the worked example", {
  result <- chi_square(example_table)
  expect_equal(result$statistic, 0.7936507936507936, tolerance = 1e-12)
  expect_equal(result$df, 1)
  expect_equal(result$p_value, pchisq(0.7936507936507936, 1, lower.tail = FALSE), tolerance = 1e-12)
})

test_that("chi_square() agrees with chisq.test() on the census", {
  tab <- provider_table(sites())
  ours <- chi_square(tab)
  theirs <- suppressWarnings(chisq.test(tab, correct = FALSE))
  expect_equal(ours$statistic, unname(theirs$statistic), tolerance = 1e-9)
  expect_equal(ours$df, unname(theirs$parameter))
  expect_equal(ours$p_value, theirs$p.value, tolerance = 1e-12)
})

test_that("Cramér's V", {
  expect_equal(cramers_v(example_table), 0.0890870806374748, tolerance = 1e-12)
  v <- cramers_v(provider_table(sites()))
  expect_equal(v, 0.745062240281175, tolerance = 1e-9)
})
