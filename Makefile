.PHONY: proto generate install-deps clean migrate migrate-up migrate-down migrate-status migrate-create


PROTO_DIR = api/proto
GEN_DIR = pkg/proto/v1
PROTO_FILE = $(PROTO_DIR)/shelduler.proto
DB_PATH = internal/migrations/data/task_scheduler.db
MIGRATIONS_DIR = internal/migrations


install-deps:
	@echo "Downloading dependencies..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "✓ Dependencies downloaded"

proto: install-deps
	@echo "Generating proto files..."
	@mkdir -p $(GEN_DIR)
	@protoc \
		--proto_path=$(PROTO_DIR) \
		--proto_path=$$(go list -f '{{ .Dir }}' -m google.golang.org/protobuf)/.. \
		--go_out=. \
		--go_opt=module=task_shelduler \
		--go-grpc_out=. \
		--go-grpc_opt=module=task_shelduler \
		$(PROTO_FILE)
	@echo "✓ Proto files generated in $(GEN_DIR)"


generate: proto
	@echo "✓ Generation completed"


clean:
	@echo "Cleaning generated files..."
	@rm -rf $(GEN_DIR)
	@echo "✓ Cleaning completed"


check-protoc:
	@which protoc > /dev/null || (echo "ERROR: protoc not installed. Install: https://grpc.io/docs/protoc-installation/" && exit 1)
	@echo "✓ protoc found: $$(protoc --version)"


all: check-protoc proto
	@echo "✓ All done!"



# ------------------------------------------------------------

migrate:
	@mkdir -p $$(dirname $(DB_PATH))
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) up

migrate-up: migrate
	@echo "✓ Migrations applied"

migrate-down:
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) down

migrate-status:
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) status

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "ERROR: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@goose -dir $(MIGRATIONS_DIR) create $(NAME) sql
	@echo "✓ Migration created: $(MIGRATIONS_DIR)/*_$(NAME).sql"

migrate-reset:
	@echo "Resetting SQLite database at $(DB_PATH)..."
	@rm -f $(DB_PATH) $(DB_PATH)-shm $(DB_PATH)-wal
	@make migrate
	@echo "✓ Database reset and migrations reapplied"




# ------------------------------------------------------------

docker-up:
	@echo "Starting Kafka and Zookeeper..."
	@docker compose up -d
	@echo "✓ Kafka is running on localhost:9092"

docker-down:
	@echo "Stopping Kafka and Zookeeper..."
	@docker compose down
	@echo "✓ Kafka stopped"

docker-logs:
	@docker compose logs -f kafka


# run-all: docker-up
# 	@echo "Waiting for Kafka to be ready..."
# 	@sleep 5
# 	@echo "✓ Ready! Now you can run:"
# 	@echo "  Terminal 1: go run ./cmd/app"
# 	@echo "  Terminal 2: go run ./cmd/worker"

