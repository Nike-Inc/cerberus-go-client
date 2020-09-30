clean:
	go clean

test:
	go test -v ./api ./auth ./cerberus ./utils

build:
	go build -o cerberus-client

.PHONY: test bootstrap build
