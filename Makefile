.PHONY: proto generate install-deps clean


PROTO_DIR = api/proto
GEN_DIR = pkg/proto/v1
PROTO_FILE = $(PROTO_DIR)/shelduler.proto


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

