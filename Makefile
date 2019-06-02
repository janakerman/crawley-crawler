.PHONY: build clean deploy

build: ## Build Lambda functions.
	GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/crawler functions/crawler/*

clean:
	rm -rf ./bin

deploy: clean build ## Build and deploy lambda functions.
	sls deploy --verbose

debug: clean ## Build a debug binary and package it up.
	GOARCH=amd64 GOOS=linux go build -gcflags='-N -l' -o bin/crawler functions/crawler/*
	if [ ! -f bin/dlv ]; then \
		GOARCH=amd64 GOOS=linux go build -o bin/dlv "-ldflags=-s -X main.Build=04834a781abd1388c21670d2c1eb49045d5f1b04" github.com/go-delve/delve/cmd/dlv; \
	fi
	sls package
	sls sam export --output ./template.yml
	sam local invoke -d 5986  -e test/events/crawlRequest.json --env-vars env.json --region eu-west-2 --debugger-path bin --debug-args -delveAPI=2

invoke:
	sls invoke  -e test/events/crawlRequest.json --env-vars env.json -f crawler
	# aws sqs send-message --queue-url https://sqs.eu-west-2.amazonaws.com/743259902374/CrawlRequestQ --message-body "{\"URL\":\"https://www.bbc.co.uk/news\"}" --region eu-west-2

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
