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
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"

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

// Subscription is a subscription.
type Subscription struct {
	CrawlID         string
	ConnectionID    string
	GatewayEndpoint string
}

// LinkRelationship represents a relationship between a page and a page it links to.
type LinkRelationship struct {
	CrawlID   string
	ParentURL string
	ChildURLs []string
}

// APIGatewayClient is an APIGatewayClient
func APIGatewayClient(endpoint string) *apigatewaymanagementapi.ApiGatewayManagementApi {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	config := aws.NewConfig().WithEndpoint(endpoint)

	return apigatewaymanagementapi.New(sess, config)
}

func unmarshalStreamImage(attribute map[string]events.DynamoDBAttributeValue, out interface{}) error {

	dbAttrMap := make(map[string]*dynamodb.AttributeValue)

	for k, v := range attribute {

		var dbAttr dynamodb.AttributeValue

		bytes, marshalErr := v.MarshalJSON()
		if marshalErr != nil {
			return marshalErr
		}

		json.Unmarshal(bytes, &dbAttr)
		dbAttrMap[k] = &dbAttr
	}

	return dynamodbattribute.UnmarshalMap(dbAttrMap, out)
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

func handleEvent(event events.DynamoDBEvent) {
	for _, record := range event.Records {
		if record.EventName == "INSERT" || record.EventName == "MODIFY" {
			rel := LinkRelationship{}
			unmarshalStreamImage(record.Change.NewImage, &rel)

			subs := getSubscribers(rel.CrawlID)
			notifySubscribers(rel, subs)
		}
	}
}

func getSubscribers(crawlID string) []Subscription {
	table := os.Getenv("CRAWL_SUB_TABLE_NAME")
	gsi := os.Getenv("CRAWL_SUB_GSI_NAME")

	// TODO: Handle paging of subscriptions???
	input := &dynamodb.QueryInput{
		TableName: &table,
		IndexName: &gsi,
		KeyConditions: map[string]*dynamodb.Condition{
			"CrawlID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(crawlID),
					},
				},
			},
		},
	}

	fmt.Printf("Querying: %#v\n", input)
	result, err := dynamoClient.Query(input)

	if err != nil {
		fmt.Println(err.Error())
		panic(fmt.Sprintf("Failed to get subscriptions"))
	}

	subscriptions := []Subscription{}
	for _, item := range result.Items {
		fmt.Printf("Item: %#v", item)

		sub := Subscription{}
		dynamodbattribute.UnmarshalMap(item, &sub)
		subscriptions = append(subscriptions, sub)
	}

	fmt.Printf("Got subscriptions: %#v\n", subscriptions)
	return subscriptions
}

func notifySubscribers(relationship LinkRelationship, subscribers []Subscription) {
	callGateway := func(endpoint, connectionID, crawlID string, payload []byte) {
		fmt.Printf("Notifying endpoint: %s payload: %s\n", endpoint, string(payload))

		apiGatewayClient := APIGatewayClient(endpoint)

		output, error := apiGatewayClient.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(connectionID),
			Data:         payload,
		})

		if error != nil {
			fmt.Println("Error posting to gateway:", error.Error())

			// TODO: Check if 410 GONE - connectionID disconnected (how?)
			// TODO: Add TTL for records so that they age off after some period longer than the connection timeout.
			fmt.Printf("Received 410 GONE error - connectionID '%s' is stale\n", connectionID)

			table := os.Getenv("CRAWL_SUB_TABLE_NAME")
			_, error := dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"ConnectionID": {
						S: aws.String(connectionID),
					},
					"CrawlID": {
						S: aws.String(crawlID),
					},
				},
				TableName: &table,
			})

			if error != nil {
				fmt.Printf("Error deleting stale connectionID: '%s' from subscription table: '%s'. Cause: %s\n", connectionID, table, error.Error())

			} else {
				fmt.Printf("Successfully deleted stale connectionID: '%s' from subscription table: '%s'\n", connectionID, table)
			}
		} else {
			fmt.Println("Posted to gateway. ", output.String())
		}
	}

	fmt.Printf("Notifying subscribers: %#v\n", subscribers)

	for _, sub := range subscribers {
		payload, _ := json.Marshal(relationship)
		callGateway(sub.GatewayEndpoint, sub.ConnectionID, sub.CrawlID, payload)
	}
}

// Handler handles DynamoDB stream events
func Handler(ctx context.Context, event events.DynamoDBEvent) (Response, error) {
	fmt.Printf("%#v\n\n", ctx)
	fmt.Printf("%#v\n\n", event)

	handleEvent(event)

	return success(""), nil
}

func main() {
	lambda.Start(Handler)
}
