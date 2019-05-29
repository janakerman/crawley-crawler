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
	ChildURL  string
}

// Crawl takes a URL and saves the child URLs it references.
// 1. Parse the HTML of a given URL.
// 2. Filter on certain conditions.
// 3. Saves the relationships to DynamoDB.
// 4. Bail out if we get too deep.
// 5. Schedule the child URLs to crawling.
func Crawl(req URLParseRequest) {
	currentURL, _ := url.Parse(req.URL)

	urls := filterURLs(ScrapeURLs(req.URL))
	saveLinkRelationships(req.CrawlID, *currentURL, urls)

	if isTooDeep(req) {
		return
	}

	scheduleChildren(req, urls)
}

func saveLinkRelationships(crawlID string, parent url.URL, children []url.URL) []LinkRelationship {
	table := os.Getenv("CRAWL_TABLE_NAME")

	rels := []LinkRelationship{}
	for _, child := range children {
		rel := LinkRelationship{
			CrawlID:   crawlID,
			ParentURL: parent.String(),
			ChildURL:  child.String(),
		}
		rels = append(rels, rel)

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

		fmt.Printf("Saved LinkRelationship: %s -> %s to table: %s\n", parent.String(), child.String(), table)
	}

	return rels
}

func filterURLs(urls []url.URL) []url.URL {
	return urls
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
	fmt.Println("Add to scrape queue")
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
