test_that("Wilson interval for the headline figure", {
  ci <- wilson_ci(4086, 7340)
  expect_named(ci, c("lower", "upper"))
  expect_equal(unname(ci), c(0.5452842329687871, 0.5680079731150852), tolerance = 1e-9)
})

test_that("Wilson matches R's own score interval", {
  for (case in list(c(4086, 7340), c(39, 73), c(12, 272), c(3, 10))) {
    ours <- unname(wilson_ci(case[1], case[2]))
    theirs <- as.numeric(prop.test(case[1], case[2], correct = FALSE)$conf.int)
    expect_equal(ours, theirs, tolerance = 1e-9)
  }
})

test_that("a higher confidence level gives a wider interval", {
  narrow <- wilson_ci(39, 73, conf = 0.90)
  wide <- wilson_ci(39, 73, conf = 0.99)
  expect_lt(wide[["lower"]], narrow[["lower"]])
  expect_gt(wide[["upper"]], narrow[["upper"]])
})

test_that("Wilson stays inside [0, 1] at the extremes; Wald doesn't", {
  expect_equal(wilson_ci(0, 10)[["lower"]], 0, tolerance = 1e-12)
  expect_lte(wilson_ci(10, 10)[["upper"]], 1 + 1e-12)
  expect_gt(wilson_ci(0, 10)[["upper"]], 0)

  # Amazon CloudFront: 695 of 698.
  expect_gt(wald_ci(695, 698)[["upper"]], 1)
  expect_lte(wilson_ci(695, 698)[["upper"]], 1)
  expect_equal(unname(wilson_ci(695, 698)), c(0.9874404207198965, 0.9985372403358771), tolerance = 1e-9)
})

test_that("Wald and Wilson agree when n is large and p is near a half", {
  expect_equal(unname(wald_ci(4086, 7340)), unname(wilson_ci(4086, 7340)), tolerance = 1e-3)
})

test_that("add_cis() adds percentage intervals around each row", {
  by <- add_cis(summarize_pq(sites(), by = "provider"))
  expect_true(all(c("lower", "upper") %in% names(by)))
  expect_true(all(by$lower <= by$percent & by$percent <= by$upper))
  cf <- by[by$provider == "Amazon CloudFront", ]
  expect_equal(c(cf$lower, cf$upper), c(98.74404207198965, 99.85372403358771), tolerance = 1e-7)
})
