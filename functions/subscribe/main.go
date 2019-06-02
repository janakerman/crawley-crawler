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
)

// DynamoSession returns a new DynamoDB client.
func DynamoSession() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return dynamodb.New(sess)
}

var dynamoClient = DynamoSession()

// Response is a response.
type Response events.APIGatewayProxyResponse

// SubscriptionRequest is a request for a subscription.
type SubscriptionRequest struct {
	CrawlID string
}

// Subscription is a subscription.
type Subscription struct {
	CrawlID         string
	ConnectionID    string
	GatewayEndpoint string
}

func subscribe(sub Subscription) Subscription {
	table := os.Getenv("CRAWL_SUB_TABLE_NAME")

	fmt.Printf("Saving subscription: '%#v' to table: '%s'\n", sub, table)

	item, err := dynamodbattribute.MarshalMap(sub)
	if err != nil {
		fmt.Println("Error marshalling Subscription:")
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

	return sub
}

func handleSubscription(event events.APIGatewayWebsocketProxyRequest) {
	req := SubscriptionRequest{}
	json.Unmarshal([]byte(event.Body), &req)

	subscribe(Subscription{
		CrawlID:         req.CrawlID,
		ConnectionID:    event.RequestContext.ConnectionID,
		GatewayEndpoint: fmt.Sprintf("https://%s/%s", event.RequestContext.DomainName, event.RequestContext.Stage),
	})
}

func unsubscribe(connectionID string) {
	table := os.Getenv("CRAWL_SUB_TABLE_NAME")
	fmt.Printf("Removing subscriptions for connectionID: '%s' from table: '%s'\n", connectionID, table)

	input := dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ConnectionID": {
				S: aws.String(connectionID),
			},
		},
		TableName: aws.String(table),
	}

	_, err := dynamoClient.DeleteItem(&input)
	if err != nil {
		fmt.Println("Error marshalling Subscription:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func handleMessage(event events.APIGatewayWebsocketProxyRequest) (Response, error) {
	route := event.RequestContext.RouteKey
	fmt.Printf("Received route '%s'", route)

	switch route {
	case "$connect":
	case "$disconnect":
		unsubscribe(event.RequestContext.ConnectionID)
	case "subscribe":
		handleSubscription(event)
	default:
		fmt.Println("Default route")
		panic("Unknown route")
	}
	return success(""), nil
}

func success(body string) Response {
	return Response{
		StatusCode: 200,
		Body:       body,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// Handler handles SQS url crawl events
func Handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (Response, error) {
	fmt.Printf("%#v\n\n", ctx)
	fmt.Printf("%#v\n\n", event)

	handleMessage(event)

	return success(""), nil
}

func main() {
	lambda.Start(Handler)
}
