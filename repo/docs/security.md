# Security Hardening & Local Deployment Notes

## TLS & Key Management
- The backend **must run over HTTPS**. Set `TLS_ENABLED=true` (default) and provide valid paths for `TLS_CERT_FILE` and `TLS_KEY_FILE` pointing to PEM cert and private key files (e.g., from `backend/certs`).
- Example self-signed cert creation:
  ```bash
  openssl req -x509 -newkey rsa:2048 -nodes \
    -keyout backend/certs/server.key \
    -out backend/certs/server.crt \
    -days 365 \
    -subj "/CN=localhost"
  ```
  These files are mounted into the Docker container (`./backend/certs:/certs:ro`).
- The server fails fast if `JWT_SECRET` or `ENCRYPTION_KEY` are missing, preventing weak defaults. Provide these via your environment (or `.env` file) before running Docker.

## IP & Transport Controls
- A global IP allowlist (`ALLOWED_IPS` CSV/CIDR) blocks all traffic except trusted LAN segments and loopback; `/health` remains exempt for health checks.
- Proxy headers (`X-Forwarded-For`, `X-Real-IP`) are ignored unless `TRUST_PROXY_HEADERS=true` is set and your ingress controller is trusted.
- CORS is constrained to `CORS_ALLOWED_ORIGINS`, avoiding wildcard origins.
- Additional security headers (HSTS, `X-Frame-Options`, `X-Content-Type-Options`, `Referrer-Policy`) are emitted via middleware.

## Encryption & Logging
- AES-256 encrypts sensitive fields (addresses, medical notes, phone numbers) at rest; the encryption key is sourced from `ENCRYPTION_KEY`.
- Rotary log files respect `LOG_FILE`, `LOG_MAX_BYTES`, and `LOG_MAX_BACKUPS`. Logs redact phone/address values and never serialize raw authorization tokens.

## Local Docker Checklist
1. Create `backend/certs/server.crt` + `backend/certs/server.key` as shown above.
2. Export secrets in your shell or `.env`:
   ```bash
   export JWT_SECRET="$(uuidgen)"
   export ENCRYPTION_KEY="$(openssl rand -hex 32)"
   ```
3. Ensure `.env` or environment includes `ALLOWED_IPS`, `TRUST_PROXY_HEADERS`, and `CORS_ALLOWED_ORIGINS` for your LAN.
4. Start services: `docker-compose up --build` (backend only starts once its secrets are present).
5. Visit [https://localhost:8443/docs](https://localhost:8443/docs) after trusting the cert in your browser.
