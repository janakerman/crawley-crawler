package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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

	urls := ScrapeURLs(req.URL)
	urls = filterURLs(
		urls,
		isSameDomain(*currentURL),
		isNotEmpty,
	)
	urls = relativeURLsToAbsolute(*currentURL, urls)

	saveLinkRelationships(req.CrawlID, *currentURL, urls)

	if isTooDeep(req) {
		return
	}

	scheduleChildren(req, urls)
}

func saveLinkRelationships(crawlID string, parent url.URL, children []url.URL) LinkRelationship {
	table := os.Getenv("CRAWL_TABLE_NAME")

	childrenString := []string{}
	for _, child := range children {
		childrenString = append(childrenString, child.String())
	}

	rel := LinkRelationship{
		CrawlID:   crawlID,
		ParentURL: parent.String(),
		ChildURLs: childrenString,
	}

	item, err := dynamodbattribute.MarshalMap(rel)
	if err != nil {
		fmt.Println("Error marshalling LinkRelationship:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(table),
	}

	_, err = dynamoClient.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("Saved LinkRelationships: %v to table: %s\n\n", rel, table)

	return rel
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

func filterURLs(urls []url.URL, predicates ...func(url url.URL) bool) []url.URL {
	filteredURLs := []url.URL{}
	for _, url := range urls {
		include := true
		for _, predicate := range predicates {
			include = include && predicate(url)
		}
		if include {
			filteredURLs = append(filteredURLs, url)
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
		scheduleForScrape(request)
	}
	return requests
}

func scheduleForScrape(req URLParseRequest) {
	fmt.Printf("Add URL '%s' to scrape queue.\n", req.URL)
}

func isTooDeep(req URLParseRequest) bool {
	max, _ := strconv.ParseUint(os.Getenv("CRAWL_MAX_DEPTH"), 10, 32)
	if req.Depth >= int(max) {
		return true
	}
	return false
}

func dynamoSession() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return dynamodb.New(sess)
}

func isCycling(crawlID string, url url.URL) bool {
	table := os.Getenv("CRAWL_TABLE_NAME")
	result, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: &table,
		Key: map[string]*dynamodb.AttributeValue{
			"CrawlID": {
				S: aws.String(crawlID),
			},
			"ParentURL": {
				S: aws.String(url.String()),
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		panic(fmt.Sprintf("Failed get record"))
	}

	item := LinkRelationship{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}

	if item.ParentURL == "" {
		return false
	}

	fmt.Printf("Already visited URL '%s' in crawl '%s'.\n", url.String(), crawlID)
	return true
}
