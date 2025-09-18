DB_IMAGE        ?= postgres:16
DB_USER         ?= postgres
DB_PASSWORD     ?= bear
DB_NAME         ?= notedb_test
DB_CONTAINER    ?= notes_test_db
DB_PORT         ?= 5432
DB_URL          ?= postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

MIGRATIONS_DIR  ?= ./migrations
FIXTURES_DIR    ?= ./test/fixtures

GOOSE_BIN       ?= goose               
GO_TEST_FLAGS   ?= ./internal/httpServer -v
FIXTURE_ACCOUNT_ID ?= 101
FIXTURE_NOTE_ID ?= 1

.PHONY: test db-up db-wait migrate fixtures db-down db-clean db-logs psql url

db-up:
	@echo ">>> Starting test Postgres container $(DB_CONTAINER) on port $(DB_PORT)"
	@docker rm -f $(DB_CONTAINER) >/dev/null 2>&1 || true
	@docker run -d --name $(DB_CONTAINER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		$(DB_IMAGE)
	sleep 2
	@$(MAKE) db-wait


db-wait:
	@echo ">>> Waiting for Postgres to accept connections..."
	@for i in $$(seq 1 60); do \
	  docker exec $(DB_CONTAINER) pg_isready -U $(DB_USER) -d $(DB_NAME) >/dev/null 2>&1 && { echo "ok"; exit 0; }; \
	  sleep 1; \
	done; \
	echo "timeout waiting for Postgres"; \
	docker logs --tail=100 $(DB_CONTAINER) || true; \
	exit 1


migrate:
	@echo ">>> Applying migrations with goose"
	@$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up


test: db-up migrate 
	@echo ">>> Running go tests with DB_URL=$(DB_URL)"
	@DB_URL="$(DB_URL)" FIXTURE_ACCOUNT_ID=$(FIXTURE_ACCOUNT_ID) FIXTURE_NOTE_ID=$(FIXTURE_NOTE_ID) go test $(GO_TEST_FLAGS)
	@$(MAKE) db-down

db-logs:
	@docker logs -f $(DB_CONTAINER)

psql:
	@docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

url:
	@echo $(DB_URL)

db-down:
	@echo ">>> Stopping and removing $(DB_CONTAINER)"
	@docker rm -f $(DB_CONTAINER) >/dev/null 2>&1 || true

db-clean: db-down