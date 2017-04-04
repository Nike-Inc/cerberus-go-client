test:
	go test -v `glide nv`
bootstrap:
	glide install --strip-vendor
build:
	go build -o cerberus

.PHONY: test bootstrap build