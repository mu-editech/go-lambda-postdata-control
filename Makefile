.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/go-lambda-post-control main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
