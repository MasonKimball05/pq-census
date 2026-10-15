test_that("every line of the file becomes one row, sorted by rank", {
  c <- census()
  expect_s3_class(c, "data.frame")
  expect_equal(nrow(c), 10000)
  expect_type(c$rank, "integer")
  expect_equal(c$rank, 1:10000)
})

test_that("columns have the right types", {
  c <- census()
  expect_type(c$domain, "character")
  expect_type(c$ok, "logical")
  expect_type(c$pq, "logical")
  expect_type(c$provider, "character")
  expect_type(c$tls_version, "character")
  expect_type(c$ms, "double")
  expect_type(c$error, "character")
})

test_that("the counts match the Go report", {
  c <- census()
  expect_equal(sum(c$ok), 7340)
  expect_equal(sum(c$pq), 4086)
  expect_equal(sum(c$error == "dns", na.rm = TRUE), 1411)
  expect_equal(sum(c$error == "reset", na.rm = TRUE), 639)
})

test_that("unreachable sites have an error and nothing else", {
  c <- census()
  down <- c[!c$ok, ]
  expect_true(all(!is.na(down$error)))
  expect_true(all(is.na(down$provider)))
  expect_true(all(is.na(down$ms)))
  expect_false(any(down$pq))
})

test_that("reachable() keeps exactly the sites that answered", {
  s <- sites()
  expect_equal(nrow(s), 7340)
  expect_true(all(s$ok))
  expect_true(all(is.na(s$error)))
  expect_false(any(is.na(s$ms)))
  expect_false(any(is.na(s$provider)))
})
