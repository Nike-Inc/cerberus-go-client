test:
	go test -v `glide nv`
bootstrap:
	glide install
build:
	go build -o cerberus

.PHONY: test bootstrap build
