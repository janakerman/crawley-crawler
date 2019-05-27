.PHONY: build clean deploy

build: ## Build Lambda functions.
	GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/crawler functions/crawler/main.go

clean:
	rm -rf ./bin

deploy: clean build ## Build and deploy lambda functions.
	sls deploy --verbose

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
