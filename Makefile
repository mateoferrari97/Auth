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
	@echo "=> Executing golangci-lint"
	@golangci-lint run ./...
.PHONY: up
up:
	@echo "=> Building and executing docker-compose"
	@docker-compose build
	@docker-compose up -d
.PHONY: migrations
migrations:
	@echo "=> Migrating sql files..."
	@docker exec -i $(id) mysql -u$(DATABASE_USER) -p$(DATABASE_PASSWORD) $(DATABASE_NAME) < cmd/app/migrations/init.sql
.PHONY: terminal
terminal:
	@echo "=> Executing interactive mode in container: $(id)"
	@docker exec -it $(id) bash
.PHONY: mysql-login
mysql-login:
	@echo "=> Login into container database..."
	@docker exec -it $(id) mysql -u$(DATABASE_USER) -p$(DATABASE_PASSWORD)