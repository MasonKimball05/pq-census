test_that("the overall rate is the headline 55.7%", {
  all <- summarize_pq(sites())
  expect_equal(nrow(all), 1)
  expect_equal(all$sites, 7340)
  expect_equal(all$pq, 4086)
  expect_equal(all$percent, 4086 / 7340 * 100)
})

test_that("by provider, largest first, ties alphabetical", {
  by <- summarize_pq(sites(), by = "provider")
  expect_equal(names(by), c("provider", "sites", "pq", "percent"))
  expect_equal(by$provider[1:4], c("Other", "Cloudflare", "Amazon CloudFront", "Google"))
  # Three providers have 272 sites each.
  expect_equal(by$provider[5:7], c("Akamai", "Amazon (ELB/S3)", "Fastly"))
  cf <- by[by$provider == "Cloudflare", ]
  expect_equal(cf$sites, 1948)
  expect_equal(cf$pq, 1891)
  expect_equal(sum(by$sites), 7340)
})

test_that("rank bands have the right labels and counts", {
  band <- rank_band(c(1, 100, 101, 1000, 1001, 10000))
  expect_s3_class(band, "factor")
  expect_equal(levels(band), c("1-100", "101-1,000", "1,001-10,000"))
  expect_equal(as.character(band),
               c("1-100", "1-100", "101-1,000", "101-1,000", "1,001-10,000", "1,001-10,000"))

  s <- sites()
  s$band <- rank_band(s$rank)
  by <- summarize_pq(s, by = "band")
  top <- by[by$band == "1-100", ]
  expect_equal(c(top$sites, top$pq), c(73, 39))
  mid <- by[by$band == "101-1,000", ]
  expect_equal(c(mid$sites, mid$pq), c(652, 361))
})
