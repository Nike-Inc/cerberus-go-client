install:
	dep ensure -v

vendor:
	go mod vendor -v

test:
	go test -v ./api ./auth ./cerberus ./utils -coverprofile=profile.out -covermode=atomic
	cat profile.out >> ../coverage.txt
	rm profile.out

clean:
	go clean
	rm -rfv vendor

.PHONY: test clean
