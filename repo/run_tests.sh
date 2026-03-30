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

echo "Running backend unit tests..."
docker-compose exec backend go test -v ./... -coverprofile=coverage.out || {
	echo "Backend unit tests failed" >&2
	exit 1
}

echo "Running backend integration tests..."
docker-compose exec backend go test -v ./integration_tests || {
	echo "Backend integration tests failed" >&2
	exit 1
}

echo "Running frontend tests..."
docker-compose exec frontend npm test --silent || {
	echo "Frontend tests failed" >&2
	exit 1
}

echo "All tests passed."
