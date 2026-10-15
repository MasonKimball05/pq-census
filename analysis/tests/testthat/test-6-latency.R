test_that("comparing every reachable site", {
  s <- sites()
  result <- latency_compare(s)
  expect_equal(result$n_pq, 4086)
  expect_equal(result$n_classical, 7340 - 4086)
  expect_equal(result$median_pq, median(s$ms[s$pq]))
  expect_equal(result$median_classical, median(s$ms[!s$pq]))
  expect_gte(result$p_value, 0)
  expect_lte(result$p_value, 1)
})

test_that("comparing within one provider", {
  s <- sites()
  result <- latency_compare(s, provider = "Cloudflare")
  expect_equal(result$n_pq + result$n_classical, 1948)
  expect_equal(result$median_pq, 170)
  expect_equal(result$median_classical, 215)
})
