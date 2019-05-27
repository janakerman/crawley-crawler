package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	// "github.com/gocolly/colly"
)

type urlParseRequest struct {
	URL     string
	CrawlID string
}

type crawlMeta struct {
	URL         string
	lastCrawlID string
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

	if crawlID == "" || crawlID == crawlMeta.lastCrawlID {
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

func getCrawlMeta(req urlParseRequest) crawlMeta {
	return crawlMeta{}
}

func scheduleForScrape(req urlParseRequest) {
	fmt.Println("Add to scrape queue")
}

func updateCrawMeta(req urlParseRequest) crawlMeta {
	return crawlMeta{}
}

func crawlPage(req urlParseRequest) []urlParseRequest {
	// Scrape URLS and add to crawl queue.
	return []urlParseRequest{}
}
