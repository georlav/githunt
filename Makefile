lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.33.0 golangci-lint run --color=always
build:
	go build -o githunt -ldflags "-X main.version=$(shell git rev-list -1 HEAD)"