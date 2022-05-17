clean:
	go clean
	rm -rfv vendor

install:
	dep ensure -v

vendor:
	go mod vendor -v

test:
	go test -v ./api ./auth ./cerberus ./utils -coverprofile=profile.out -covermode=atomic
	cat profile.out >> ../coverage.txt
	rm profile.out

build:
	go build -o cerberus-client

.PHONY: test bootstrap build
