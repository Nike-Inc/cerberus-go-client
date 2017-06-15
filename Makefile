test:
	go test -v `glide nv`
bootstrap:
	glide install
build:
	go build -o cerberus-client

.PHONY: test bootstrap build
