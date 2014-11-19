build:	fmt
	GOPATH=$(shell pwd) go build -o bin/todos src/main.go

fmt:
	gofmt -w `find ./ -name '*.go'`
