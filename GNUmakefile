default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	go tool golangci-lint run

generate:
	go tool tfplugindocs generate --provider-dir . -provider-name piano

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install generate
