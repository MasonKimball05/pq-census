test_that("the model is a logistic regression on TLS 1.3 sites", {
  fit <- fit_rank_model(sites())
  expect_s3_class(fit, "glm")
  expect_equal(fit$family$family, "binomial")
  expect_equal(nobs(fit), 6205)
  expect_true("log10(rank)" %in% names(coef(fit)))
})

test_that("Other is the reference provider", {
  fit <- fit_rank_model(sites())
  terms <- names(coef(fit))
  expect_false("providerOther" %in% terms)
  expect_true("providerCloudflare" %in% terms)
  expect_false("providerGitHub" %in% terms)
  providers <- sort(sub("^provider", "", grep("^provider", terms, value = TRUE)))
  expect_equal(providers, c("Akamai", "Amazon (ELB/S3)", "Amazon CloudFront", "Cloudflare", "Fastly",
                            "Google", "Microsoft Azure", "Netlify", "Unknown", "Vercel"))
})

test_that("odds ratios are exponentiated coefficients with intervals around them", {
  fit <- fit_rank_model(sites())
  or <- odds_ratios(fit)
  expect_equal(names(or), c("term", "odds_ratio", "lower", "upper"))
  expect_false("(Intercept)" %in% or$term)
  expect_equal(nrow(or), length(coef(fit)) - 1)
  expect_equal(or$odds_ratio, unname(exp(coef(fit)[or$term])), tolerance = 1e-12)
  expect_true(all(or$lower < or$odds_ratio & or$odds_ratio < or$upper))
  expect_true(all(or$lower > 0))
})
