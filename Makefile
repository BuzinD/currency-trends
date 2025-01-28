.PHONY: run, start, cloneEnv, build-app, migrate, create-migration
n=?
include app/env/db.env
MIGRATE_CMD=docker run --rm -i -v ./app/migrations:/migrations/migrations --network currency docker.io/library/go-migration:0.0.1 /migrations/migrate -path=/migrations/migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${POSTGRES_DB}?sslmode=disable"
export

cloneEnv:
	cp app/env/.env.example app/env/.env
	cp app/env/db.env.example app/env/db.env
	cp app/env/okx.env.example app/env/okx.env
run:
	cd app && go run cmd/main.go
stop:
	docker compose -f ./docker/docker-compose.yml stop
start:
	docker compose -f ./docker/docker-compose.yml up -d
start-db:
	docker compose -f ./docker/docker-compose.yml up currency-db -d
build-app:
	docker build -t go-currency:0.0.1 -f docker/go/Dockerfile .
build-go-base:
	docker build -t go-base:0.0.1 -f docker/go-base/Dockerfile .
build-migration:
	docker build -t go-migration:0.0.1 -f docker/go-migration/Dockerfile .
create-migration:
	docker run --rm -i -v ./app/migrations:/migrations/migrations --network currency docker.io/library/go-migration:0.0.1 /migrations/migrate create -ext sql -dir migrations $(n)
migrate:
	${MIGRATE_CMD} up
migrate-down:
	${MIGRATE_CMD} down
tidy:
	cd app && go mod tidy
