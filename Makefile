## This is a currency trend service
.PHONY: run, start, cloneEnv, build-app, migrate, create-migration
n=?
include app/env/db.env
MIGRATE_CMD=docker run --rm -i -v ./app/migrations:/migrations/migrations --network currency docker.io/library/go-migration:0.0.1 /migrations/migrate -path=/migrations/migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${POSTGRES_DB}?sslmode=disable"
export

help: ## Show this help.
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST)
first-run-dev: cloneEnv build-postgres-db start-db build-go-base build-migration migrate run ## prepare envs, build images for db, migrations, run their and run data-fetcher
cloneEnv: ## copy examples to working env files (without overwriting)
	cp -n app/env/.env.example app/env/.env
	cp -n app/env/db.env.example app/env/db.env
	cp -n app/env/okx.env.example app/env/okx.env
run: ## run data-fetcher service
	export DB_HOST=127.0.0.1 &&	export DB_PORT=15432 && cd app && go run cmd/main.go
stop: ## stop docker containers
	docker compose -f ./docker/docker-compose.yml stop
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
	docker run --rm -i -v ./app/migrations:/migrations/migrations --network currency docker.io/library/go-migration:0.0.1 /migrations/migrate create -ext sql -dir migrations $(n)
migrate: ## migrate up
	${MIGRATE_CMD} up
migrate-down: ## migrate down
	${MIGRATE_CMD} down
tidy: ## go mod tidy
	cd app && go mod tidy
