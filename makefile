##### GLOBAL #####
GOOSE_BIN              ?= goose
WAIT_MAX               ?= 60

##### NOTES (notes-svc) #####
NOTES_DB_IMAGE         ?= postgres:16
NOTES_DB_USER          ?= postgres
NOTES_DB_PASSWORD      ?= bear
NOTES_DB_NAME          ?= notedb_test
NOTES_DB_CONTAINER     ?= notes_test_db
NOTES_DB_PORT_HOST     ?= 5442
NOTES_DB_PORT_CONT     ?= 5432
NOTES_DB_URL           ?= postgres://$(NOTES_DB_USER):$(NOTES_DB_PASSWORD)@localhost:$(NOTES_DB_PORT_HOST)/$(NOTES_DB_NAME)?sslmode=disable

NOTES_MIGRATIONS_DIR   ?= ./migrations/notes-svc
NOTES_FIXTURES_DIR     ?= ./internal/notes-svc/httpServer/testdata/fixtures
GO_TEST_NOTES_FLAGS    ?= -v ./internal/notes-svc/... --count=1 -p=1

##### GATEWAY (auth, http gateway) #####
GATEWAY_DB_IMAGE       ?= postgres:16
GATEWAY_DB_USER        ?= postgres
GATEWAY_DB_PASSWORD    ?= bear
GATEWAY_DB_NAME        ?= gatewaydb_test
GATEWAY_DB_CONTAINER   ?= gateway_test_db
GATEWAY_DB_PORT_HOST   ?= 5441
GATEWAY_DB_PORT_CONT   ?= 5432
GATEWAY_DB_URL         ?= postgres://$(GATEWAY_DB_USER):$(GATEWAY_DB_PASSWORD)@localhost:$(GATEWAY_DB_PORT_HOST)/$(GATEWAY_DB_NAME)?sslmode=disable

GATEWAY_MIGRATIONS_DIR ?= ./migrations/gateway
GATEWAY_FIXTURES_DIR   ?= ./test/fixtures
GO_TEST_GATEWAY_FLAGS  ?= -v ./internal/gateway/httpServer -count=1 -failfast

##### PHONY #####
.PHONY: test-all \
        notes-db-up notes-db-wait notes-migrate notes-test notes-db-logs notes-psql notes-url notes-db-down notes-db-clean \
        gateway-db-up gateway-db-wait gateway-migrate gateway-test gateway-db-logs gateway-psql gateway-url gateway-db-down gateway-db-clean

##### NOTES TARGETS #####
notes-db-up:
	@echo ">>> Starting NOTES Postgres $(NOTES_DB_CONTAINER) on :$(NOTES_DB_PORT_HOST)"
	@docker rm -f $(NOTES_DB_CONTAINER) >/dev/null 2>&1 || true
	@docker run -d --name $(NOTES_DB_CONTAINER) \
		-e POSTGRES_PASSWORD=$(NOTES_DB_PASSWORD) \
		-e POSTGRES_DB=$(NOTES_DB_NAME) \
		-p $(NOTES_DB_PORT_HOST):$(NOTES_DB_PORT_CONT) \
		$(NOTES_DB_IMAGE) >/dev/null
	@$(MAKE) notes-db-wait

notes-db-wait:
	@echo ">>> Waiting for NOTES Postgres to accept connections..."
	@for i in $$(seq 1 $(WAIT_MAX)); do \
	  docker exec $(NOTES_DB_CONTAINER) pg_isready -U $(NOTES_DB_USER) -d $(NOTES_DB_NAME) >/dev/null 2>&1 && { echo "ok"; exit 0; }; \
	  sleep 1; \
	done; \
	echo "timeout waiting for NOTES Postgres"; \
	docker logs --tail=100 $(NOTES_DB_CONTAINER) || true; \
	exit 1

notes-migrate:
	@echo ">>> Applying NOTES migrations with goose"
	@$(GOOSE_BIN) -dir $(NOTES_MIGRATIONS_DIR) postgres "$(NOTES_DB_URL)" up

notes-test: notes-db-up notes-migrate
	@echo ">>> Running NOTES tests with DB_URL=$(NOTES_DB_URL)"
	@NOTES_DB_URL="$(NOTES_DB_URL)" \
		go test $(GO_TEST_NOTES_FLAGS)
	@$(MAKE) notes-db-down

notes-db-logs:
	@docker logs -f $(NOTES_DB_CONTAINER)

notes-psql:
	@docker exec -it $(NOTES_DB_CONTAINER) psql -U $(NOTES_DB_USER) -d $(NOTES_DB_NAME)

notes-url:
	@echo $(NOTES_DB_URL)

notes-db-down:
	@echo ">>> Stopping and removing $(NOTES_DB_CONTAINER)"
	@docker rm -f $(NOTES_DB_CONTAINER) >/dev/null 2>&1 || true

notes-db-clean: notes-db-down

##### GATEWAY TARGETS #####
gateway-db-up:
	@echo ">>> Starting GATEWAY Postgres $(GATEWAY_DB_CONTAINER) on :$(GATEWAY_DB_PORT_HOST)"
	@docker rm -f $(GATEWAY_DB_CONTAINER) >/dev/null 2>&1 || true
	@docker run -d --name $(GATEWAY_DB_CONTAINER) \
		-e POSTGRES_PASSWORD=$(GATEWAY_DB_PASSWORD) \
		-e POSTGRES_DB=$(GATEWAY_DB_NAME) \
		-p $(GATEWAY_DB_PORT_HOST):$(GATEWAY_DB_PORT_CONT) \
		$(GATEWAY_DB_IMAGE) >/dev/null
	@$(MAKE) gateway-db-wait

gateway-db-wait:
	@echo ">>> Waiting for GATEWAY Postgres to accept connections..."
	@for i in $$(seq 1 $(WAIT_MAX)); do \
	  docker exec $(GATEWAY_DB_CONTAINER) pg_isready -U $(GATEWAY_DB_USER) -d $(GATEWAY_DB_NAME) >/dev/null 2>&1 && { echo "ok"; exit 0; }; \
	  sleep 1; \
	done; \
	echo "timeout waiting for GATEWAY Postgres"; \
	docker logs --tail=100 $(GATEWAY_DB_CONTAINER) || true; \
	exit 1

gateway-migrate:
	@echo ">>> Applying GATEWAY migrations with goose"
	@$(GOOSE_BIN) -dir $(GATEWAY_MIGRATIONS_DIR) postgres "$(GATEWAY_DB_URL)" up

gateway-test: gateway-db-up gateway-migrate
	@echo ">>> Running GATEWAY tests with DB_URL=$(GATEWAY_DB_URL)"
	@GATEWAY_DB_URL="$(GATEWAY_DB_URL)" go test $(GO_TEST_GATEWAY_FLAGS)
	@$(MAKE) gateway-db-down

gateway-db-logs:
	@docker logs -f $(GATEWAY_DB_CONTAINER)

gateway-psql:
	@docker exec -it $(GATEWAY_DB_CONTAINER) psql -U $(GATEWAY_DB_USER) -d $(GATEWAY_DB_NAME)

gateway-url:
	@echo $(GATEWAY_DB_URL)

gateway-db-down:
	@echo ">>> Stopping and removing $(GATEWAY_DB_CONTAINER)"
	@docker rm -f $(GATEWAY_DB_CONTAINER) >/dev/null 2>&1 || true

gateway-db-clean: gateway-db-down

##### ALL #####
test-all:
	@$(MAKE) notes-test
	@$(MAKE) gateway-test