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

## Key Provisioning & Rotation (Guidance)

- Provision secrets via a secure mechanism (environment, secret manager) and avoid committing them to source control.
- Recommended production practice:
  - `JWT_SECRET`: rotate annually or on suspicion of compromise. Use a 32+ character random secret (e.g., generated from `/dev/urandom` or a secrets manager).
  - `ENCRYPTION_KEY`: must be exactly 32 bytes (AES-256). Rotate with care: maintain a key versioning metadata so previously encrypted records remain decryptable until re-encrypted with the new key.
  - Store key rotation metadata (version, activation time) outside the repository and provide the active key via the environment.
- Backup/rotate steps:
  1. Generate new key: `openssl rand -hex 32` (store securely).
  2. Deploy the new key alongside the old key in a controlled rollout window.
  3. Run a re-encryption job that reads records with the old key and re-encrypts them with the new key, preserving availability.
  4. Remove old key only after successful re-encryption and validation.

Include key-rotation in your operational runbooks and restrict access to key materials to a small set of operations/admin users.

## Sensitive Data Handling in Responses & Logs

- Do NOT include full decrypted medical or session notes in API responses. Provide only a short summary and a presence indicator where necessary.
- Ensure structured logs never contain decrypted notes or full authorization tokens. The application logger redacts known sensitive keys, but review any custom logging to avoid accidental leaks.
- When implementing export or scheduled reports, write encrypted fields to disk in encrypted form or write only summaries unless explicit operational approval is provided.
