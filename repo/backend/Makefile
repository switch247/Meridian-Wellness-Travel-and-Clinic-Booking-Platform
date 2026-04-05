.PHONY: test test-unit test-integration test-local

# Run all Go tests in backend (unit + integration)
test: test-unit test-integration

# Run unit tests located under ./tests/
test-unit:
	@echo "Running backend unit tests..."
	go test ./tests/...

# Run API/integration tests under ./tests/API_tests/
test-integration:
	@echo "Running backend integration tests..."
	RUN_INTEGRATION_TESTS=true BASE_URL=http://localhost:8443 go test ./tests/API_tests/...

# Local-first verification path (no Docker daemon required).
# DB-backed tests auto-skip when DATABASE_URL is not set.
test-local:
	@echo "Running local backend verification suite..."
	go test ./internal/... ./tests/...
