package main

import (
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestFirstAllowedLink(t *testing.T) {
	cases := []struct {
		body    string
		wantURL string
		wantOK  bool
	}{
		{"check out https://github.com/foo/bar", "https://github.com/foo/bar", true},
		{"https://www.github.com/a", "https://www.github.com/a", true}, // subdomain allowed
		{"https://x.com/user", "https://x.com/user", true},
		{"https://evil.com/payload", "", false},      // not whitelisted
		{"https://sub.evil.com/payload", "", false},  // subdomain not whitelisted
		{"https://github.com.evil.com/x", "", false}, // suffix spoof rejected
		{"no links here", "", false},
		{"ftp://github.com/x", "", false},                                              // non-http scheme rejected
		{"https://github.com/a then https://evil.com/b", "https://github.com/a", true}, // first allowed wins
	}
	for _, tc := range cases {
		got, ok := firstAllowedLink(tc.body)
		if ok != tc.wantOK || got != tc.wantURL {
			t.Errorf("firstAllowedLink(%q) = (%q, %v), want (%q, %v)",
				tc.body, got, ok, tc.wantURL, tc.wantOK)
		}
	}
}

func TestIsAllowed(t *testing.T) {
	allowed := []string{"github.com", "www.github.com", "api.github.com", "x.com", "twitter.x.com"}
	for _, h := range allowed {
		if !isAllowed(h) {
			t.Errorf("expected %q to be allowed", h)
		}
	}
	disallowed := []string{"evil.com", "github.com.evil.com", "notgithub.com", "xcom", "foo.github.com.evil.com"}
	for _, h := range disallowed {
		if isAllowed(h) {
			t.Errorf("expected %q to be disallowed", h)
		}
	}
}

func TestExtractOG(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`<html><head>
		<meta property="og:title" content="Repo Title">
		<meta name="og:description" content="A repo description">
		<meta property="og:image" content="https://github.com/img.png">
		<meta property="og:url" content="https://github.com/foo/bar">
	</head></html>`))
	if err != nil {
		t.Fatal(err)
	}
	meta := extractOG(doc, "https://github.com/foo/bar")
	if meta.Title != "Repo Title" {
		t.Errorf("title = %q, want Repo Title", meta.Title)
	}
	if meta.Description != "A repo description" {
		t.Errorf("description = %q, want A repo description", meta.Description)
	}
	if meta.Image != "https://github.com/img.png" {
		t.Errorf("image = %q, want github img", meta.Image)
	}
	if meta.URL != "https://github.com/foo/bar" {
		t.Errorf("url = %q, want original link", meta.URL)
	}
}

func TestFormatPreview(t *testing.T) {
	got := formatPreview(ogMeta{Title: "T", Description: "D", URL: "https://github.com/a"})
	want := `<a href="https://github.com/a">T</a>
<blockquote>D</blockquote>`
	if got != want {
		t.Errorf("formatPreview = %q, want %q", got, want)
	}
}

func TestPlainPreview(t *testing.T) {
	got := plainPreview(ogMeta{Title: "T", Description: "D", URL: "https://github.com/a"})
	want := "T\nD\nhttps://github.com/a"
	if got != want {
		t.Errorf("plainPreview = %q, want %q", got, want)
	}
}

func TestPlainPreviewEmptyFields(t *testing.T) {
	tests := []struct {
		meta ogMeta
		name string
	}{
		{ogMeta{URL: "https://x.com/a"}, "title only"},
		{ogMeta{Description: "desc", URL: "https://x.com/a"}, "description only"},
		{ogMeta{Title: "T", Description: "D"}, "no url"},
	}
	for _, tc := range tests {
		got := plainPreview(tc.meta)
		if got == "" {
			t.Errorf("plainPreview(%s) returned empty string", tc.name)
		}
	}
}

func TestCleanURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://github.com/foo/bar?utm_source=twitter", "https://github.com/foo/bar"},
		{"https://github.com/foo/bar?fbclid=abc&ref=x", "https://github.com/foo/bar?ref=x"},
		{"https://github.com/foo/bar", "https://github.com/foo/bar"},
		{"https://github.com/foo/bar?ref=homepage", "https://github.com/foo/bar?ref=homepage"}, // not a tracking param
		{"https://x.com/user/status/123?s=20&t=foo", "https://x.com/user/status/123"},
		{"https://x.com/user/status/123?utm_campaign=x_social", "https://x.com/user/status/123"},
		{"https://x.com/user/status/123?foo=bar", "https://x.com/user/status/123?foo=bar"}, // not tracking
	}
	for _, tc := range cases {
		got := cleanURL(tc.in)
		if got != tc.want {
			t.Errorf("cleanURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCleanURLPreservesPathAndQuery(t *testing.T) {
	u, err := url.Parse("https://github.com/foo/bar?utm_source=x&q=1")
	if err != nil {
		t.Fatal(err)
	}
	cleaned := cleanParsed(u)
	expected := "https://github.com/foo/bar?q=1"
	if cleaned != expected {
		t.Errorf("cleanParsed = %q, want %q", cleaned, expected)
	}
}
