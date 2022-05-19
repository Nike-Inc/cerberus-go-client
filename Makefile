test:
	go test -v ./api ./auth ./cerberus ./utils -coverprofile=profile.out -covermode=atomic
	cat profile.out >> coverage.txt
	rm profile.out

clean:
	go clean

.PHONY: test clean 
