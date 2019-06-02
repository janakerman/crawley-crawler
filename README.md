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


## TODO

* Batch write of items.
* Correct following of sub/parent domains.
* Unit tests on depth etc.
* Checking of previously parsed nodes? Add additional information? At the moment DynamoDB will simply be acting as a queue - kind of pointless.
* Add TTL to crawl records. Add overall crawl meta to table when started (initial API gateway lambda)