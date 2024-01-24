build:	test
	go build -o bin/todos main.go

test:	resource-constants
	go test ./...

resource-constants: fmt
	mkdir -p bin
	go build -o bin/resource-constants utils/resource-constants.go
	if [ ! -e "resources" ]; then mkdir resources; fi
	bin/resource-constants --base_dir $(shell pwd)/ui/ > resources/constants.go

fmt:
	gofmt -w `find ./ -name '*.go'`

clean:
	rm -r bin || true
	rm -r resources || true