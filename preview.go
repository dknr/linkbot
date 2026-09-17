package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// allowedDomains is the whitelist of hosts whose links the bot will preview.
// A link is previewed only when its host equals one of these or is a
// subdomain of one (e.g. www.github.com).
var allowedDomains = []string{"github.com", "x.com"}

// urlRe matches http(s) URLs in a message body.
var urlRe = regexp.MustCompile(`https?://[^\s<]+`)

// ogMeta holds the Open Graph tags pulled from a page.
type ogMeta struct {
	Title       string
	Description string
	Image       string
	URL         string
}

// firstAllowedLink scans body for the first URL whose host is in the
// whitelist. It returns ok=false when no allowed link is found.
func firstAllowedLink(body string) (string, bool) {
	for _, match := range urlRe.FindAllString(body, -1) {
		u, err := url.Parse(match)
		if err != nil {
			continue
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			continue
		}
		if !isAllowed(u.Host) {
			continue
		}
		return u.String(), true
	}
	return "", false
}

// isAllowed reports whether host is on the whitelist, matching the domain
// itself or any subdomain of it.
func isAllowed(host string) bool {
	host = strings.ToLower(host)
	for _, d := range allowedDomains {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

// fetchOG downloads the page and extracts its Open Graph meta tags. Redirects
// are followed only when the destination host is still on the whitelist, so a
// whitelisted page cannot bounce us onto a disallowed host.
func fetchOG(ctx context.Context, link string) (ogMeta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return ogMeta{}, err
	}
	req.Header.Set("User-Agent", "linkbot/1.0")
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 5 {
				return fmt.Errorf("too many redirects")
			}
			if !isAllowed(req.URL.Host) {
				return fmt.Errorf("redirect to disallowed host %q", req.URL.Host)
			}
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return ogMeta{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ogMeta{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	doc, err := html.Parse(io.LimitReader(resp.Body, 2<<20)) // cap the body at 2 MiB
	if err != nil {
		return ogMeta{}, err
	}
	return extractOG(doc, link), nil
}

// extractOG walks the parsed document and reads the og:* meta tags.
func extractOG(doc *html.Node, link string) ogMeta {
	meta := ogMeta{URL: link}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, a := range n.Attr {
				switch a.Key {
				case "property", "name":
					name = strings.ToLower(a.Val)
				case "content":
					content = a.Val
				}
			}
			switch name {
			case "og:title":
				if meta.Title == "" {
					meta.Title = content
				}
			case "og:description":
				if meta.Description == "" {
					meta.Description = content
				}
			case "og:image":
				if meta.Image == "" {
					meta.Image = content
				}
			case "og:url":
				if meta.URL == "" {
					meta.URL = content
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return meta
}

// formatPreview renders the preview as a plain-text reply: title, description,
// then the resolved URL.
func formatPreview(meta ogMeta) string {
	var b strings.Builder
	if meta.Title != "" {
		b.WriteString(meta.Title)
		b.WriteString("\n")
	}
	if meta.Description != "" {
		b.WriteString(meta.Description)
		b.WriteString("\n")
	}
	b.WriteString(meta.URL)
	return b.String()
}
