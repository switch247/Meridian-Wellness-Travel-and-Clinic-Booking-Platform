Concurrency Runner

This small Go program fires concurrent `POST /api/v1/bookings/holds` requests to exercise the booking concurrency logic.

Usage (requires backend running and reachable):

1. Build:

```bash
cd tools/concurrency-runner
go build -o concurrency-runner
```

2. Run (example):

```bash
./concurrency-runner -base https://localhost:8443 -concurrency 10 -insecure
```

Flags:
- `-base` API base URL (default https://localhost:8443)
- `-concurrency` number of concurrent clients
- `-package`, `-host`, `-room`, `-duration` configure the booking payload
- `-slot` RFC3339 slot start (default now+4h)
- `-insecure` skip TLS verification (useful for self-signed certs)

Notes:
- The program registers per-client users and logs them in to obtain tokens.
- Run this inside the Docker network or on a machine that can reach the backend endpoint.
- It's a test harness — interpret results and re-run as needed to validate anti-oversell behavior.
