install:
	dep ensure -v

vendor:
	go mod vendor -v

test:
	rm -f ../coverage.txt
	go test -v ./api ./auth ./cerberus ./utils -coverprofile=profile.out -covermode=atomic
	cat profile.out >> ../coverage.txt
	rm -f profile.out

#Create html coverage report
cover: test
	rm -f cover.html
	go tool cover -html ../coverage.txt -o cover.html

lint:
	golangci-lint run -v

clean:
	go clean
	rm -rfv vendor

.PHONY: test clean
