.PHONY: generate
generate:
	oapi-codegen --config codegen.yaml openapi.yaml

.PHONY: build
build: generate
	go build -o bin/server ./cmd/server

.PHONY: run
run: generate
	go run ./cmd/server

.PHONY: test
test: generate
	go test ./...

.PHONY: clean
clean:
	rm -f internal/server/http/*.gen.go