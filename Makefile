.PHONY: lint test vendor clean

export GO111MODULE=on

lint:
	golangci-lint run

test:
	go test -v -cover ./...

vendor:
	go mod vendor

clean:
	rm -rf ./vendor