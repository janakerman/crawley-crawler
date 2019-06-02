package main

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

// Visitor is callback when visiting an anchor tag.
type Visitor func(url.URL)

// ScrapeURLs requests a given page and returns all URLs referenced from it matching.
// filter - Only return URLs matching this filter.
func ScrapeURLs(requestURL string) []url.URL {
	fmt.Println("Requesting url: ", requestURL)
	resp, _ := http.Get(requestURL)
	fmt.Println("Got response: ", resp.StatusCode)
	urls := []url.URL{}

	visitURLs(resp, func(url url.URL) {
		urls = append(urls, url)
	})

	resp.Body.Close()
	return urls
}

func visitURLs(response *http.Response, visitor Visitor) {
	tokenizer := html.NewTokenizer(response.Body)
	finished := false

	for !finished {
		tt := tokenizer.Next()

		switch {
		case tt == html.ErrorToken:
			finished = true
		case tt == html.StartTagToken:
			token := tokenizer.Token()

			isAnchor := token.Data == "a"
			if isAnchor {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						href := attr.Val
						url, error := url.Parse(href)
						if error != nil {
							fmt.Printf("Error parsing <a> href '%s'\n:", href)
							fmt.Println(error.Error())
							break
						}
						visitor(*url)
					}
				}
			}
		}
	}
}
