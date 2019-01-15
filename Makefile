clean:
	go clean
	rm -rfv vendor

install:
	dep ensure -v

test:
	go test -v ./api ./auth ./cerberus ./utils

build:
	go build -o cerberus-client

.PHONY: test bootstrap build
