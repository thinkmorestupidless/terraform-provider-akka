BINARY_NAME=terraform-provider-akka
PLUGIN_DIR=~/.terraform.d/plugins/registry.terraform.io/thinkmorestupidless/akka/0.1.0/$$(go env GOOS)_$$(go env GOARCH)

default: build

build:
	go build -o $(BINARY_NAME) .

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./internal/provider/... -v -timeout 30m

generate:
	go generate ./...
	tfplugindocs generate

install: build
	mkdir -p $(PLUGIN_DIR)
	cp $(BINARY_NAME) $(PLUGIN_DIR)/

vet:
	go vet ./...

fmt:
	gofmt -s -w .

fmt-check:
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "The following files need formatting (run 'make fmt' to fix):"; \
		gofmt -s -l .; \
		exit 1; \
	fi

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

tfmt:
	terraform fmt -recursive

tfmt-check:
	terraform fmt -recursive -check

tflint:
	tflint --recursive

# Run all checks without modifying files (suitable for CI)
check: fmt-check vet lint tfmt-check

.PHONY: build test testacc generate install vet fmt fmt-check lint tidy tfmt tfmt-check tflint check
