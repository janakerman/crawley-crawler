package main

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

// ScrapeURLs asdf
func ScrapeURLs(url string) []*url.URL {
	fmt.Println("Requesting url: ", url)
	resp, _ := http.Get(url)
	fmt.Println("Got response: ", resp.StatusCode)
	urls := collectUrls(resp)
	resp.Body.Close()
	return urls
}

func collectUrls(response *http.Response) []*url.URL {
	tokenizer := html.NewTokenizer(response.Body)

	urls := []*url.URL{}

	for {
		tt := tokenizer.Next()

		switch {
		case tt == html.ErrorToken:
			return urls
		case tt == html.StartTagToken:
			token := tokenizer.Token()

			isAnchor := token.Data == "a"
			if isAnchor {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						fmt.Println("We found a link!")
						fmt.Println(attr.Val)
					}
				}
			}
		}
	}
}
