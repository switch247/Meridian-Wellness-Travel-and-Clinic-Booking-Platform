.PHONY: test test-unit test-integration

# Run all Go tests in backend (unit + integration)
test: test-unit test-integration

# Run unit tests located under ./tests/
test-unit:
	@echo "Running backend unit tests..."
	go test ./tests/...

# Run integration tests under ./integration_tests/
test-integration:
	@echo "Running backend integration tests..."
	go test ./integration_tests/...
