.PHONY: build fmt lint test tools

build:
	go build -o withings-cli ./cmd/withings

fmt:
	gofumpt -w $$(go list -f '{{.Dir}}' ./...)

lint:
	golangci-lint run ./...

test:
	go test ./...

tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
