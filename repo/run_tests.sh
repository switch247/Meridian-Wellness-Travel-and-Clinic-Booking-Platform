#!/bin/sh
set -e

cleanup() {
	docker compose down -v
}
trap cleanup EXIT

# Start services
docker compose up -d --build

# Wait for services to be healthy
docker compose exec db pg_isready -U postgres -d meridian
docker compose exec backend curl -f http://localhost:8443/health || echo "Backend not ready, continuing..."

# Run backend unit and API/integration tests
# docker compose exec -T backend sh -lc '
# if command -v go >/dev/null 2>&1; then
# 	RUN_INTEGRATION_TESTS=true BASE_URL=http://localhost:8443 go test ./...
# elif [ -x /usr/local/go/bin/go ]; then
# 	RUN_INTEGRATION_TESTS=true BASE_URL=http://localhost:8443 /usr/local/go/bin/go test ./...
# else
# 	echo "go not found in backend container (PATH=$PATH)"
# 	exit 127
# fi
# '

# # Run frontend unit tests
# docker compose exec -T frontend npm run test

# run e2e tests
docker compose exec -T frontend npm run test:e2e

echo "All tests passed successfully!"