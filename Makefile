BINARY := server
SWAG   := $(shell go env GOPATH)/bin/swag

.PHONY: build run swag tidy vet fmt clean swag-install

build: ## Compile the server binary
	go build -o $(BINARY) ./cmd/server

run: ## Run the server from source
	go run ./cmd/server

swag: ## Regenerate the Swagger spec in docs/
	$(SWAG) init -g cmd/server/main.go -o docs --parseInternal

swag-install: ## Install the swag CLI
	go install github.com/swaggo/swag/cmd/swag@latest

tidy: ## Sync go.mod/go.sum
	go mod tidy

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go sources
	gofmt -w .

clean: ## Remove build output
	rm -f $(BINARY)
