package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

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

func getLinkRelationship(crawlID string, url url.URL) LinkRelationship {
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

	return item
}

// DynamoSession returns a new DynamoDB client.
func DynamoSession() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return dynamodb.New(sess)
}
