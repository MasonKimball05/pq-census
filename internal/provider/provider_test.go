package provider

import (
	"net/http"
	"testing"
)

func TestDetect(t *testing.T) {
	cases := []struct {
		want    string
		headers map[string]string
	}{
		{"Cloudflare", map[string]string{"Server": "cloudflare"}},
		{"Cloudflare", map[string]string{"Cf-Ray": "8a1b2c3d4e5f-DFW"}},
		{"Vercel", map[string]string{"Server": "Vercel", "X-Vercel-Id": "iad1::abc"}},
		{"GitHub", map[string]string{"Server": "GitHub.com", "Via": "1.1 varnish", "X-Served-By": "cache-dfw-1"}},
		{"Amazon CloudFront", map[string]string{"Via": "1.1 abc.cloudfront.net (CloudFront)"}},
		{"Fastly", map[string]string{"X-Served-By": "cache-bhm-KBHM1234"}},
		{"Akamai", map[string]string{"Server": "AkamaiGHost"}},
		{"Google", map[string]string{"Server": "gws"}},
		{"Microsoft Azure", map[string]string{"X-Azure-Ref": "0abc"}},
		{"Other", map[string]string{"Server": "nginx"}},
		{"Other", nil},
	}
	for _, c := range cases {
		h := http.Header{}
		for k, v := range c.headers {
			h.Set(k, v)
		}
		if got := Detect(h); got != c.want {
			t.Errorf("Detect(%v) = %q, want %q", c.headers, got, c.want)
		}
	}
}
