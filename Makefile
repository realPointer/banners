LOCAL_BIN:=$(CURDIR)/bin

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help

compose-up: ### Run docker-compose
	docker-compose up --build -d && docker-compose logs -f
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down

docker-rm-volume: ### Remove docker volume
	docker volume rm banners_pg-data banners_redis-data
.PHONY: docker-rm-volume

install-all: ### Install all tools
	make install-linter
	make install-migrate
	make install-swag
.PHONY: install-all

install-linter: ### Install golangci-lint
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2
.PHONY: install-linter

lint: ### Check by golangci linter
	$(LOCAL_BIN)/golangci-lint run
.PHONY: lint

test: ### Run unit tests
	go test -v ./internal/...
.PHONY: test

integration-test: ### Run integration-test
	go clean -testcache && go test -v ./integration-tests/...
.PHONY: integration-test

install-migrate: ### Install migrate
	GOBIN=$(LOCAL_BIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0
.PHONY: install-migrate

migrate-create: ### Create migration file
	$(LOCAL_BIN)/migrate create -ext sql -dir migrations "banners"
.PHONY: migrate-create

install-swag: ### Install swag
	GOBIN=$(LOCAL_BIN) go install github.com/swaggo/swag/cmd/swag@v1.16.3

swag: ### Generate swag docs
	swag init -g cmd/app/main.go
.PHONY: swag