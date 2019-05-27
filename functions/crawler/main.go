package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	// "github.com/gocolly/colly"
)

var dynamoClient = dynamoSession()

type urlParseRequest struct {
	URL     string
	CrawlID string
}

type crawlMeta struct {
	URL         string
	LastCrawlID string
}

// Handler handles SQS url crawl events
func Handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		handleRecord(message)
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}

func handleRecord(message events.SQSMessage) {
	urlParseReq := parseRecord(message)
	crawlMeta := getCrawlMeta(urlParseReq)
	crawlID := urlParseReq.CrawlID

	if crawlID == "" || crawlID == crawlMeta.LastCrawlID {
		fmt.Println("Page already crawled for CrawlID:", crawlID)
		return
	}

	fmt.Println("Crawling page for CrawlID:", crawlID)

	// Locking?
	scheduleForScrape(urlParseReq)
	updateCrawMeta(urlParseReq)

	crawlPage(urlParseReq)
}

func parseRecord(message events.SQSMessage) urlParseRequest {
	req := urlParseRequest{}
	s := message.Body
	json.Unmarshal([]byte(s), &req)

	if req.CrawlID == "" {
		req.CrawlID = uuid.New().String()
	}

	return req
}

func getCrawlMeta(req urlParseRequest) *crawlMeta {
	table := os.Getenv("CRAWL_TABLE_NAME")
	url := req.URL

	result, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]*dynamodb.AttributeValue{
			"URL": {
				S: aws.String(url),
			},
		},
	})
	if err != nil {
		fmt.Println("Error getting CrawlMeta. CrawlId: ", req.CrawlID)
		fmt.Println(err.Error())
		return nil
	}

	meta := crawlMeta{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &meta)

	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}

	fmt.Printf("Got CrawlMeta: %v, from table: '%s'\n", meta, table)

	return &meta
}

func scheduleForScrape(req urlParseRequest) {
	fmt.Println("Add to scrape queue")
}

func updateCrawMeta(req urlParseRequest) crawlMeta {
	table := os.Getenv("CRAWL_TABLE_NAME")
	meta := crawlMeta{
		URL:         req.URL,
		LastCrawlID: req.CrawlID,
	}

	fmt.Printf("Saving CrawlMeta: %v, to table: '%s'\n", meta, table)

	item, err := dynamodbattribute.MarshalMap(meta)
	if err != nil {
		fmt.Println("Error marshalling CrawlMeta:")
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

	fmt.Printf("Saved CrawlMeta: %v\n", meta)

	return meta
}

func crawlPage(req urlParseRequest) []urlParseRequest {
	// Scrape URLS and add to crawl queue.

	return []urlParseRequest{}
}

func dynamoSession() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return dynamodb.New(sess)
}
