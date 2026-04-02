#!/bin/sh
set -e

# Start services
docker compose up -d --build

# Wait for services to be healthy
docker compose exec db pg_isready -U postgres -d meridian
docker compose exec backend curl -f http://localhost:8443/health || echo "Backend not ready, continuing..."

# Run backend tests
docker compose exec backend go test ./...

# Run frontend tests
docker compose exec frontend npm run test


# Stop services
docker compose down

echo "All tests passed successfully!"