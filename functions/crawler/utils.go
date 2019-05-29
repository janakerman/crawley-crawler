package main

import (
	"log"
	"net/url"
)

// NormaliseURL removes all but host & path from the URL.
func NormaliseURL(fullURL string) string {
	url, err := url.ParseRequestURI(fullURL)
	if err != nil {
		log.Fatal("Failed to parse URL.", err)
	}

	url.User = nil
	url.ForceQuery = false
	url.RawQuery = ""
	url.Fragment = ""

	return url.String()
}
