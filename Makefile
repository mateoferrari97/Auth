.PHONY: all
all: dependencies fmt linter
.PHONY: dependencies
dependencies:
	@echo "=> Executing go mod tidy for ensure dependencies..."
	@go mod tidy
.PHONY: fmt
fmt:
	@echo "=> Executing go fmt..."
	@go fmt ./...
.PHONY: linter
linter:
	@echo "=> Executing golangci-lint $(if $(FLAGS), with flags: $(FLAGS))..."
	@golangci-lint run ./... $(FLAGS)