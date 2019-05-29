package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
)

var dynamoClient = dynamoSession()

// URLParseRequest represents a request to crawl a specific page.
type URLParseRequest struct {
	URL     string
	CrawlID string
	Depth   int
}

// Handler handles SQS url crawl events
func Handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		Crawl(parseRecord(message))
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}

func parseRecord(message events.SQSMessage) URLParseRequest {
	req := URLParseRequest{}
	s := message.Body
	json.Unmarshal([]byte(s), &req)

	if req.CrawlID == "" {
		req.CrawlID = uuid.New().String()
	}

	if req.Depth == 0 {
		req.Depth = 1
	}

	return req
}
