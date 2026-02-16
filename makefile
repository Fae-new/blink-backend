.PHONY: generate-module test-prod test-prod-setup test-prod-run test-prod-cleanup check-health fetch-staging-db
-include .env
export

DB_URL="postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"

dev:
	@set -a && . ./.env && set +a && air

dev-up:
	docker compose up -d

dev-down:
	docker compose down

test-prod: dev-up wait-for-db test-prod-setup test-prod-run check-health

wait-for-db:
	@echo "Waiting for PostgreSQL to be ready..."
	@for i in $$(seq 1 30); do \
		if docker exec postman-runner-postgres pg_isready -h localhost -p 5432 -U dev > /dev/null 2>&1; then \
			echo "\033[32m✓ PostgreSQL is ready\033[0m"; \
			break; \
		fi; \
		echo "Waiting for PostgreSQL to start... ($$i/30)"; \
		sleep 1; \
	done

test-prod-run:
	docker compose --env-file .env -f docker-compose.prod.yml up -d

check-health:
	@echo "Waiting for service to start..."
	@for i in $$(seq 1 30); do \
		if curl -s http://localhost:8080/health > /dev/null; then \
			echo "\033[32m✨ Service is healthy and ready!\033[0m"; \
			echo "\033[36m🚀 Production setup is working correctly\033[0m"; \
			echo "\033[35m💫 You're good to deploy!\033[0m"; \
			break; \
		fi; \
		echo "Waiting for service to start... ($$i/30)"; \
		sleep 1; \
	done

test-prod-cleanup:
	@echo "Restoring original environment..."
	@if [ -f .env.backup ]; then \
		mv .env.backup .env; \
	fi
	docker compose -f docker-compose.prod.yml down
	make dev-down

migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir ./migrations create $$name sql

migrate-up:
	goose -dir ./migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir ./migrations postgres "$(DB_URL)" down

migrate-status:
	goose -dir ./migrations postgres "$(DB_URL)" status

generate-module:
	@if [ "$(name)" = "" ]; then \
		echo "Usage: make generate-module name=<module_name>"; \
		exit 1; \
	fi
	@go run generate_module.go $(name)

dump-staging-db:
	@echo "Creating db-dumps directory if it doesn't exist..."
	@mkdir -p db-dumps
	@echo "Dumping staging database..."
	@docker run --rm -v "$(PWD)/db-dumps:/dumps" postgres:17.6 bash -c "PGPASSWORD=$(DB_PASSWORD) pg_dump -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f /dumps/postman_runner_dump.sql"
	@echo "✅ Staging database dumped to db-dumps/postman_runner_dump.sql"

apply-db-dump:
	@echo "Ensuring local PostgreSQL is running..."
	@if ! docker ps | grep -q postman-runner-postgres; then \
		echo "Starting Docker containers..."; \
		docker compose up -d postgres; \
		echo "Waiting for PostgreSQL to start..."; \
		sleep 5; \
	fi
	@echo "Dropping existing database..."
	@docker exec -i postman-runner-postgres psql -U dev -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'postman_runner_db' AND pid <> pg_backend_pid();" postgres
	@docker exec -i postman-runner-postgres psql -U dev -c "DROP DATABASE IF EXISTS postman_runner_db;" postgres
	@echo "Recreating empty database..."
	@docker exec -i postman-runner-postgres psql -U dev -c "CREATE DATABASE postman_runner_db;" postgres
	@echo "Applying database dump to local instance..."
	@docker exec -i postman-runner-postgres psql -U dev -d postman_runner_db < db-dumps/postman_runner_dump.sql
	@echo "✅ Successfully reset and applied database dump"

fetch-staging-db: dump-staging-db apply-db-dump

dump-staging-schema:
	@echo "Creating db-schema directory if it doesn't exist..."
	@mkdir -p db-schema
	@echo "Dumping staging database schema only..."
	@docker run --rm -v "$(PWD)/db-schema:/dumps" postgres:17.6 bash -c "PGPASSWORD=$(DB_PASSWORD) pg_dump --schema-only --no-owner --no-privileges -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f /dumps/postman_runner_schema.sql"
	@echo "✅ Staging database schema dumped to db-schema/postman_runner_schema.sql"

apply-staging-schema:
	@echo "Ensuring local PostgreSQL is running..."
	@if ! docker ps | grep -q postman-runner-postgres; then \
		echo "Starting Docker containers..."; \
		docker compose up -d postgres; \
		echo "Waiting for PostgreSQL to start..."; \
		sleep 5; \
	fi
	@echo "Dropping existing database..."
	@docker exec -i postman-runner-postgres psql -U dev -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'postman_runner_db' AND pid <> pg_backend_pid();" postgres
	@docker exec -i postman-runner-postgres psql -U dev -c "DROP DATABASE IF EXISTS postman_runner_db;" postgres
	@echo "Recreating empty database..."
	@docker exec -i postman-runner-postgres psql -U dev -c "CREATE DATABASE postman_runner_db;" postgres
	@echo "Applying schema dump to local instance..."
	@docker exec -i postman-runner-postgres psql -U dev -d postman_runner_db < db-schema/postman_runner_schema.sql
	@echo "✅ Successfully reset and applied schema dump"

fetch-staging-schema: dump-staging-schema apply-staging-schema

test:
	go test ./... -v -cover

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout 5m

build-agent:
	@echo "Building agent binaries..."
	@mkdir -p downloads
	@cd agent && make build-mac && mv bin/blink-agent-darwin-amd64 ../downloads/blink-agent-darwin-amd64
	@cd agent && make build-mac-arm && mv bin/blink-agent-darwin-arm64 ../downloads/blink-agent-darwin-arm64
	@cd agent && make build-windows && mv bin/blink-agent-windows-amd64.exe ../downloads/blink-agent-windows-amd64.exe
	@cd agent && make build-linux && mv bin/blink-agent-linux-amd64 ../downloads/blink-agent-linux-amd64
	@echo "✅ Agent binaries built and moved to downloads/"
