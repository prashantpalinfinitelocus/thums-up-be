.PHONY: build dev clean up down logs docs rebuild rebuild-no-cache db-up db-down db-setup

APP_NAME := thums_up_backend

build:
	go build -o main .

dev:
	go run main.go server

clean:
	go clean
	rm -f main

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f --tail=100

docs:
	swag init -g cmd/server.go

rebuild:
	docker compose build
	docker compose up -d

rebuild-no-cache:
	docker compose build --no-cache
	docker compose up -d

build-subscriber:
	docker compose -f docker-compose.subscriber.yml build

up-subscriber:
	docker compose -f docker-compose.subscriber.yml up -d

down-subscriber:
	docker compose -f docker-compose.subscriber.yml down

db-up:
	docker compose -f docker-compose.db.yml up -d

db-down:
	docker compose -f docker-compose.db.yml down

db-setup:
	./scripts/setup-database.sh

db-reset:
	docker compose -f docker-compose.db.yml down -v
	docker compose -f docker-compose.db.yml up -d

strapi-dev:
	cd strapi && npm run dev

strapi-build:
	cd strapi && npm run build

strapi-install:
	cd strapi && npm install

