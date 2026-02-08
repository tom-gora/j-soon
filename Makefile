BINARY_NAME := jsoon
BUILD_DIR := bin

.PHONY: all build prepare_test_data test_unit test_emit test_validate test_limit test_file test_lookahead test_pipeline test_info test_all clean

all: build

build:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/j-soon/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

prepare_test_data:
	@echo "Preparing relative test data..."
	@./test_data/prepare_test_calendar_data.sh

test_unit:
	@echo "Running Go Unit Tests..."
	@go test -v ./...

test_emit: build prepare_test_data
	@echo "Running Emit Test..."
	@cd test_data && ./test_emit_notifications.sh 10
	@$(MAKE) clean

test_validate: build prepare_test_data
	@echo "Running Validation Test..."
	@cd test_data && ./test_validate_json.sh 7
	@$(MAKE) clean

test_limit: build prepare_test_data
	@echo "Running Limit Test..."
	@cd test_data && ./test_limit_events.sh
	@$(MAKE) clean

test_file: build prepare_test_data
	@echo "Running File Output Test..."
	@cd test_data && ./test_file_output.sh
	@$(MAKE) clean

test_lookahead: build prepare_test_data
	@echo "Running Number of Lookahead Days Test..."
	@./test_data/test_lookahead.sh
	@$(MAKE) clean

test_pipeline: build prepare_test_data
	@echo "Running Pipeline (stdin) Test..."
	@cd test_data && ./test_pipeline.sh
	@$(MAKE) clean

test_info: build
	@echo "Running Info Flags Test..."
	@cd test_data && ./test_info_flags.sh
	@$(MAKE) clean

# test_all runs everything without cleaning between steps
test_all: build prepare_test_data
	@echo "Running All Tests..."
	@go test -v ./...
	@cd test_data && ./test_validate_json.sh 7
	@cd test_data && ./test_limit_events.sh
	@cd test_data && ./test_file_output.sh
	@./test_data/test_lookahead.sh
	@cd test_data && ./test_pipeline.sh
	@cd test_data && ./test_info_flags.sh
	@cd test_data && ./test_emit_notifications.sh 1
	@$(MAKE) clean

clean:
	@rm -rf $(BUILD_DIR)
