build:	test
	GOPATH=$(shell pwd) go build -o bin/todos src/main.go

test:	resource-constants
	GOPATH=$(shell pwd) go test `find src -name '*_test.go' | sed 's/^src\///' | uniq | xargs dirname`

resource-constants: fmt
	go build -o bin/resource-constants utils/resource-constants.go
	if [ ! -e "src/resources" ]; then mkdir src/resources; fi
	bin/resource-constants --base_dir $(shell pwd)/src/ui/ > src/resources/constants.go

fmt:
	gofmt -w `find ./ -name '*.go'`

clean:
	rm -r bin || true
	rm -r src/resources || true

# The following rule copies the locally built binary into our publicly readable
# download location. This will fail for anyone who is not on the core todo-tracks team.
publish: build
	gsutil cp -a public-read bin/todos gs://todo-track-bin/todos
