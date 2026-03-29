#!/usr/bin/env sh
set -eu

docker-compose up -d --build

docker-compose exec backend ./run_tests.sh
