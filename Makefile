## This is a currency trend service
.PHONY: run, start, cloneEnv, build-app, migrate, create-migration
n=?

APP_FETCHER_DIR=data-fetcher
MIGRATIONS_PATH=$(shell pwd)/$(APP_FETCHER_DIR)/migrations

include $(APP_FETCHER_DIR)/env/db.env
MIGRATE_CMD=docker run --rm -i -v $(MIGRATIONS_PATH):/migrations/migrations --network currency docker.io/library/go-migration:0.0.1 /migrations/migrate -path=/migrations/migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${POSTGRES_DB}?sslmode=disable"
export

help: ## Show this help.
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST)
first-run-dev: cloneEnv build-postgres-db start-db build-go-base build-migration migrate run ## prepare envs, build images for db, migrations, run their and run data-fetcher
cloneEnv: ## copy examples to working env files (without overwriting)
	cp --update=none $(APP_FETCHER_DIR)/env/.env.example $(APP_FETCHER_DIR)/env/.env || true
	cp --update=none $(APP_FETCHER_DIR)/env/db.env.example $(APP_FETCHER_DIR)/env/db.env || true
	cp --update=none $(APP_FETCHER_DIR)/env/okx.env.example $(APP_FETCHER_DIR)/env/okx.env || true
run: ## run data-fetcher service
	export DB_HOST=127.0.0.1 &&	export DB_PORT=15432 && cd $(APP_FETCHER_DIR) && go run cmd/main.go
build: ## build a data-fetcher app
	cd data-fetcher && go build -o cmd/data-fetcher cmd/main.go
test-env-up: ## up test env and db
	export DB_HOST=currency-db-test && docker compose -f ./docker/docker-compose-test.yml up -d
test: ## run tests (run 'make test-env-up' before)
	export DB_PORT=25432 && export DB_HOST=localhost && cd $(APP_FETCHER_DIR) && go test ./... -v
test-env-down:
	docker compose -f ./docker/docker-compose-test.yml down

down: ## stop docker containers
	docker compose -f ./docker/docker-compose.yml down
start: ## start docker containers
	docker compose -f ./docker/docker-compose.yml up -d
start-db: ## start docker container for db
	docker compose -f ./docker/docker-compose.yml up currency-db -d
build-app: ## build docker image for data-fetcher service
	docker build -t go-currency:0.0.1 -f docker/go/Dockerfile .
build-go-base: ## build docker image
	docker build -t go-base:0.0.1 -f docker/go-base/Dockerfile .
build-migration: ## build docker image
	docker build -t go-migration:0.0.1 -f docker/go-migration/Dockerfile .
build-postgres-db: ## build db docker image
	docker build -t currency-db:0.0.1 -f docker/postgres/Dockerfile .
create-migration: ## create a new migration. Use syntax: make create-migration n=<name of migration>
	docker run --rm -i -v ./$(APP_FETCHER_DIR)/migrations:/migrations/migrations --network currency docker.io/library/go-migration:0.0.1 /migrations/migrate create -ext sql -dir migrations $(n)
migrate: ## migrate up
	${MIGRATE_CMD} up
migrate-down: ## migrate down
	${MIGRATE_CMD} down
tidy: ## go mod tidy
	cd $(APP_FETCHER_DIR) && go mod tidy
