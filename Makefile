BINARY_NAME := cal-event-notifier
BUILD_DIR := bin

.PHONY: all build test_unit test_emit test_validate test_limit test_file test_lookahead test_pipeline test_config test_info test_all clean

all: build

build:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) JSON-from-iCal.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

test_unit:
	@echo "Running Go Unit Tests..."
	@go test -v ./...

test_emit: build
	@echo "Running Emit Test..."
	@cd test_data && ./test_emit_notifications.sh 5
	@$(MAKE) clean

test_validate: build
	@echo "Running Validation Test..."
	@cd test_data && ./test_validate_json.sh 7
	@$(MAKE) clean

test_limit: build
	@echo "Running Limit Test..."
	@cd test_data && ./test_limit_events.sh
	@$(MAKE) clean

test_file: build
	@echo "Running File Output Test..."
	@cd test_data && ./test_file_output.sh
	@$(MAKE) clean

test_lookahead: build
	@echo "Running Number of Lookahead Days Test..."
	@cd test_data && ./test_lookahead.js
	@$(MAKE) clean

test_pipeline: build
	@echo "Running Pipeline (stdin) Test..."
	@cd test_data && ./test_pipeline.sh
	@$(MAKE) clean

test_config: build
	@echo "Running Config Junk Test..."
	@cd test_data && ./test_config_junk.sh
	@$(MAKE) clean

test_info: build
	@echo "Running Info Flags Test..."
	@cd test_data && ./test_info_flags.sh
	@$(MAKE) clean

# test_all runs everything without cleaning between steps
test_all: build
	@echo "Running All Tests..."
	@go test -v ./...
	@cd test_data && ./test_validate_json.sh 7
	@cd test_data && ./test_limit_events.sh
	@cd test_data && ./test_file_output.sh
	@cd test_data && ./test_lookahead.js
	@cd test_data && ./test_pipeline.sh
	@cd test_data && ./test_config_junk.sh
	@cd test_data && ./test_info_flags.sh
	@cd test_data && ./test_emit_notifications.sh 1
	@$(MAKE) clean

clean:
	@rm -rf $(BUILD_DIR)
