// Package provider guesses which CDN or hosting platform served a response,
// from its headers. It's a heuristic: a site behind two layers (say, Fastly in
// front of S3) is attributed to the outermost one, which is the one that
// terminated TLS and so the one that chose the key exchange.
package provider

import (
	"net/http"
	"strings"
)

type rule struct {
	name  string
	match func(h http.Header, server string) bool
}

func has(h http.Header, key string) bool { return h.Get(key) != "" }

func contains(h http.Header, key, sub string) bool {
	return strings.Contains(strings.ToLower(h.Get(key)), sub)
}

// Order matters: the first match wins, so more specific platforms that run on
// top of a CDN (Vercel, Netlify) come before generic signals.
var rules = []rule{
	{"Cloudflare", func(h http.Header, s string) bool { return s == "cloudflare" || has(h, "Cf-Ray") }},
	{"Vercel", func(h http.Header, s string) bool { return s == "vercel" || has(h, "X-Vercel-Id") }},
	{"Netlify", func(h http.Header, s string) bool {
		return strings.HasPrefix(s, "netlify") || has(h, "X-Nf-Request-Id")
	}},
	{"GitHub", func(h http.Header, s string) bool { return s == "github.com" }},
	{"Amazon CloudFront", func(h http.Header, s string) bool { return has(h, "X-Amz-Cf-Id") || contains(h, "Via", "cloudfront") }},
	{"Akamai", func(h http.Header, s string) bool {
		return strings.HasPrefix(s, "akamai") || has(h, "X-Akamai-Transformed") || has(h, "Akamai-Grn") || has(h, "X-Akamai-Request-Id")
	}},
	{"Fastly", func(h http.Header, s string) bool {
		return has(h, "X-Fastly-Request-Id") || has(h, "Fastly-Restarts") || strings.HasPrefix(h.Get("X-Served-By"), "cache-")
	}},
	{"Microsoft Azure", func(h http.Header, s string) bool { return has(h, "X-Azure-Ref") || has(h, "X-Msedge-Ref") }},
	{"Google", func(h http.Header, s string) bool {
		return s == "gws" || s == "esf" || s == "google frontend" || s == "gse" || strings.HasPrefix(s, "sffe") || contains(h, "Via", "google")
	}},
	{"Amazon (ELB/S3)", func(h http.Header, s string) bool {
		return strings.HasPrefix(s, "awselb") || s == "amazons3" || has(h, "X-Amz-Request-Id")
	}},
}

// Detect returns the provider name, or "Other" when nothing matches.
func Detect(h http.Header) string {
	server := strings.ToLower(strings.TrimSpace(h.Get("Server")))
	for _, r := range rules {
		if r.match(h, server) {
			return r.name
		}
	}
	return "Other"
}
