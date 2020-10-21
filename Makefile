clean:
	go clean
	rm -rfv vendor

install:
	dep ensure -v

test:
	go test -v ./api ./auth ./cerberus ./utils
	go test -v v3/api v3/auth v3/cerberus v3/utils

build:
	go build -o cerberus-client

.PHONY: test bootstrap build
