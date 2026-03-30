#!/usr/bin/env sh
set -eu

# Bring up services
docker-compose up -d --build

echo "Waiting up to 60s for backend health endpoint..."
ready=0
for i in $(seq 1 60); do
	if curl -kfsS https://localhost:8443/health >/dev/null 2>&1; then
		ready=1
		break
	fi
	sleep 1
done

if [ "$ready" -ne 1 ]; then
	echo "Backend did not become healthy within 60s. Showing backend logs for debugging:"
	docker-compose logs --no-color backend || true
	exit 1
fi

echo "Running backend tests (all packages)..."
docker-compose exec backend go test -v ./... -coverprofile=coverage.out || {
    echo "Backend tests failed" >&2
    exit 1
}

echo "Running repository tests under ./tests/... (integration tests are skipped by default)"
docker-compose exec backend go test -v ./tests/... || {
    echo "Backend tests under ./tests failed" >&2
    exit 1
}

echo "Running frontend tests..."
docker-compose exec frontend npm test --silent || {
	echo "Frontend tests failed" >&2
	exit 1
}

echo "All tests passed."
