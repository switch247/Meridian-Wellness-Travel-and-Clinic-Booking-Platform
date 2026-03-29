# Security Hardening

## Transport Security
- HTTPS is enabled by default (`TLS_ENABLED=true`).
- Backend fails fast if `TLS_CERT_FILE` or `TLS_KEY_FILE` is missing.
- Default compose uses:
  - cert: `/certs/server.crt`
  - key: `/certs/server.key`

## Secret Key Requirements
- `JWT_SECRET` and `ENCRYPTION_KEY` must be provided through environment variables; the backend refuses to start without them.
- Choose a high-entropy `JWT_SECRET` (32+ characters) and a 32-byte `ENCRYPTION_KEY` (hex/base64) that is rotated via your deployment pipeline.

## IP Allowlist
- Global allowlist is enforced for all routes except `/health`.
- Configure via `ALLOWED_IPS` (comma-separated IP or CIDR values).
- Proxy headers are ignored by default (`TRUST_PROXY_HEADERS=false`).

## CORS and Security Headers
- CORS uses explicit origins from `CORS_ALLOWED_ORIGINS`.
- Response hardening headers:
  - Strict-Transport-Security
  - X-Content-Type-Options
  - X-Frame-Options
  - Referrer-Policy

## Local Certificate Generation
The repository includes development certificates in `backend/certs` for one-command startup.
For your own certs:

```bash
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
  -keyout backend/certs/server.key \
  -out backend/certs/server.crt \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
```

## LAN Deployment
- Add your LAN subnet to `ALLOWED_IPS` (for example `192.168.1.0/24`).
- Distribute trusted certificates to client kiosks/browsers to avoid TLS trust errors.
