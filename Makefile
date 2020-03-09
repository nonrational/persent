default:
	go build

install:
	go install

test:
	go test -v -race -coverprofile=coverage.out -covermode=atomic

coverage: test
	go tool cover -html=coverage.out

.PHONY: default
