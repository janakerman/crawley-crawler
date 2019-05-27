# CRAWLEY-CRAWLER

A serverless web crawler.

## Setup

Build & deploy to AWS using the Serverless Framework. Assumes AWS credentials set up for environment.

```
make deploy
```

Run a test event through the hosted function.

```
sls invoke -f crawler --path test/events/crawlRequest.json --env CRAWL_TABLE_NAME=CrawlTable
```
