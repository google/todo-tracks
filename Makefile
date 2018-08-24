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

# The following rule copies the locally built binary into our publicly readable
# download location. This will fail for anyone who is not on the core todo-tracks team.
publish: build
	gsutil cp -a public-read ${GOPATH:~}/bin/todos gs://todo-track-bin/todos
