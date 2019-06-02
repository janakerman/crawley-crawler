package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

// LinkRelationship represents a relationship between a page and a page it links to.
type LinkRelationship struct {
	CrawlID   string
	ParentURL string
	ChildURLs []string
}

// Crawl takes a URL and saves the child URLs it references.
// 1. Parse the HTML of a given URL.
// 2. Filter on certain conditions.
// 3. Saves the relationships to DynamoDB.
// 4. Bail out if we get too deep.
// 5. Schedule the child URLs to crawling.
func Crawl(req URLParseRequest) {
	currentURL, _ := url.Parse(req.URL)

	if isCycling(req.CrawlID, *currentURL) {
		return
	}

	filteredURLs := []url.URL{}
	filterURLs(
		ScrapeURLs(req.URL),
		&filteredURLs,
		isSameDomain(*currentURL),
		isDupelicate(&filteredURLs),
		isNotEmpty,
	)
	filteredURLs = relativeURLsToAbsolute(*currentURL, filteredURLs)

	saveLinkRelationships(req.CrawlID, *currentURL, filteredURLs)

	if isTooDeep(req) {
		return
	}

	scheduleChildren(req, filteredURLs)
}

func relativeURLsToAbsolute(parent url.URL, urls []url.URL) []url.URL {
	fixedURLs := []url.URL{}
	for _, url := range urls {
		if url.Host == "" {
			url.Scheme = parent.Scheme
			url.Host = parent.Host
		}
		fixedURLs = append(fixedURLs, url)
	}
	return fixedURLs
}

func isDupelicate(urls *[]url.URL) func(url url.URL) bool {
	return func(url url.URL) bool {
		for _, otherURL := range *urls {
			if url == otherURL {
				return false
			}
		}
		return true
	}
}

func isNotEmpty(url url.URL) bool {
	return url.Scheme != "" && url.Host != "" || url.Path != ""
}

func isSameDomain(parent url.URL) func(url url.URL) bool {
	// TODO: Check contains instead.
	return func(url url.URL) bool {
		if url.Path == "/accessibility/" {
			fmt.Println("")
		}
		return url.Host == "" || parent.Host == url.Host
	}
}

func filterURLs(urls []url.URL, filteredURLs *[]url.URL, predicates ...func(url url.URL) bool) *[]url.URL {
	for _, url := range urls {
		include := true
		for _, predicate := range predicates {
			include = include && predicate(url)
		}
		if include {
			*filteredURLs = append(*filteredURLs, url)
		} else {
			fmt.Printf("URL '%s' ignored. Failed to match predicate.\n", url.String())
		}
	}
	return filteredURLs
}

func scheduleChildren(req URLParseRequest, urls []url.URL) []URLParseRequest {
	requests := []URLParseRequest{}
	for _, url := range urls {
		request := URLParseRequest{
			URL:     url.String(),
			CrawlID: req.CrawlID,
			Depth:   req.Depth + 1,
		}
		requests = append(requests, request)
	}

	scheduleForScrape(requests)

	return requests
}

func isTooDeep(req URLParseRequest) bool {
	max, _ := strconv.ParseUint(os.Getenv("CRAWL_MAX_DEPTH"), 10, 32)
	if req.Depth >= int(max) {
		return true
	}
	return false
}

func isCycling(crawlID string, url url.URL) bool {
	rel := getLinkRelationship(crawlID, url)

	if rel.ParentURL == "" {
		return false
	}

	fmt.Printf("Already visited URL '%s' in crawl '%s'.\n", url.String(), crawlID)
	return true
}
