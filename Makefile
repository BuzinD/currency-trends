.PHONY: run, start, cloneEnv, build-app

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