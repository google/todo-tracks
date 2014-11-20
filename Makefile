build:	fmt
	GOPATH=$(shell pwd) go build -o bin/todos src/main.go

#TODO: Add a tests rule.
resource-constants: fmt
	go build -o bin/resource-constants utils/resource-constants.go
	bin/resource-constants --base_dir $(shell pwd) > src/resources.go

fmt:
	gofmt -w `find ./ -name '*.go'`

clean:
	rm -r bin || true
	rm src/resources.go || true