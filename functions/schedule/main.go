package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// CrawlRequest is a request.
type CrawlRequest struct {
	RootURL string
	CrawlID string
}

// URLParseRequest represents a request to crawl a specific page.
type URLParseRequest struct {
	URL     string
	CrawlID string
	Depth   int
}

// Response is a response.
type Response events.APIGatewayProxyResponse

func sqsClient() *sqs.SQS {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return sqs.New(sess)
}

func response(body string, statusCode int) Response {
	return Response{
		StatusCode: statusCode,
		Body:       body,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

func handleEvent(event events.APIGatewayProxyRequest) (Response, error) {
	queueURL := os.Getenv("CRAWL_QUEUE_URL")
	request := CrawlRequest{}
	json.Unmarshal([]byte(event.Body), &request)

	fmt.Printf("Received CrawlRequest: %#v\n", request)

	if request.CrawlID == "" || request.RootURL == "" {
		return response("Missing values", 400), nil
	}

	message := URLParseRequest{
		URL:     request.RootURL,
		CrawlID: request.CrawlID,
		Depth:   1,
	}

	json, _ := json.Marshal(message)
	jsonString := string(json)

	fmt.Printf("Posting initial URL parse to parse queue: %#v\n", message)

	sqsClient().SendMessage(&sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: &jsonString,
	})

	return response("", 200), nil
}

// Handler handles.
func Handler(ctx context.Context, event events.APIGatewayProxyRequest) (Response, error) {
	fmt.Printf("%#v\n\n", ctx)
	fmt.Printf("%#v\n\n", event)

	handleEvent(event)

	return response("", 200), nil
}

func main() {
	lambda.Start(Handler)
}
