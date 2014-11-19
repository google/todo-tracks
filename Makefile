build:	fmt
	GOPATH=$(shell pwd) go build -o bin/todos src/main.go

fmt:
	gofmt -w `find ./ -name '*.go'`

#TODO: Add a rule for embedding our HTML/Javascript files in Go constants.
#TODO: Add a tests rule.
