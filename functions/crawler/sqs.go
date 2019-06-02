package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func scheduleForScrape(requests []URLParseRequest) {
	queueURL := os.Getenv("CRAWL_QUEUE_URL")

	entries := []*sqs.SendMessageBatchRequestEntry{}
	for _, req := range requests {
		bytes, _ := json.Marshal(req)
		json := string(bytes)

		fmt.Printf("Adding message id: '%s' message body: '%s' to queue: '%s\n", req.URL, json, queueURL)

		batchID := urlToBatchID(req.URL)
		entries = append(entries, &sqs.SendMessageBatchRequestEntry{
			Id:          &batchID,
			MessageBody: &json,
		})

		if len(entries) == 10 {
			sendBatch(queueURL, entries)
			entries = []*sqs.SendMessageBatchRequestEntry{}
		}
	}
	sendBatch(queueURL, entries)
}

func sendBatch(queueURL string, reqs []*sqs.SendMessageBatchRequestEntry) {
	results, err := sqsClient().SendMessageBatch(&sqs.SendMessageBatchInput{
		QueueUrl: &queueURL,
		Entries:  reqs,
	})

	if err != nil {
		fmt.Println("Error posting requests to SQS", err)
		return
	}

	for _, res := range results.Successful {
		fmt.Printf("Successfully put message on queue. BatchID: '%s', MessageId: '%s' on queue: '%s'\n", *res.Id, *res.MessageId, queueURL)
	}

	for _, res := range results.Failed {
		fmt.Printf("Failed to put message '%s' on queue: '%s'\n", *res.Id, *res.Message)
	}
}

func sqsClient() *sqs.SQS {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return sqs.New(sess)
}

func urlToBatchID(url string) string {
	reg, err := regexp.Compile("[:/.]*")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(url, "")
}
