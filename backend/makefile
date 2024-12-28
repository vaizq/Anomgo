TEST_DB_COMPOSE_FILE=./docker/testdb.yml
TEST_DSN=postgresql://testuser:testpass@localhost:5432/testdb?sslmode=disable
DEV_DSN=postgresql://anom:secretpass@localhost:5432/ltdb?sslmode=disable


# Util
.PHONY: wait 
wait:
	sleep 3

# Testing
.PHONY: test-db-up
test-db-up:
	docker compose -f $(TEST_DB_COMPOSE_FILE) up -d

.PHONY: test-db-down
test-db-down:
	docker compose -f $(TEST_DB_COMPOSE_FILE) down

.PHONY: test-db-migrate-up
test-db-migrate-up:
	migrate -path ./db/migrations -database $(TEST_DSN) up

.PHONY: test-db-migrate-down
test-db-migrate-down:
	migrate -path ./db/migrations -database $(TEST_DSN) down

.PHONY: test-db-clean
test-db-clean:
	docker-compose -f $(TEST_DB_COMPOSE_FILE) down -v
	@echo "Cleaned up Docker volumes."

.PHONY: tests-run
tests-run:
	go test -count=1 -v ./internal/model ./internal/auth

.PHONY: ci-run
ci-run: test-db-clean test-db-up wait test-db-migrate-up tests-run

# Development
.PHONY: build-css
build-css:
	rm -f ./ui/static/styles/main.css
	cat ./ui/css/* >> ./ui/static/styles/main.css

.PHONY: run 
run:
	go run ./cmd/store --dsn=$(DEV_DSN)

.PHONY: admin-run 
admin-run:
	go run ./cmd/admin --dsn=$(DEV_DSN) --addr=localhost:4469


.PHONY: seed-run
seed-run:
	go run ./cmd/seed -dsn=$(DEV_DSN)

.PHONY: hotreload-run
hotreload-run:
	nodemon -w "./*" -i "./ui/static/styles/main.css" -e "go, html, css, js" -x "make build-css && make dev-run" --signal SIGTERM
