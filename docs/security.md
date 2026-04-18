# Security Notes

This document summarizes core security controls implemented in the backend.

## Authentication and Session
- Local username/password authentication
- JWT bearer tokens
- Account lockout after failed attempts, with configurable threshold and duration

## Authorization
- Route-level permission middleware
- Role-to-permission mapping in backend middleware
- Object ownership checks in handlers/repository
- Location/tenant scope checks for location-bound operations

## Data Protection
- Password hashing with bcrypt
- Application-level encryption for sensitive fields
- Masked values in user-facing responses where applicable
- Sensitive attribute redaction in logs

## Transport and Perimeter
- Configurable TLS startup with fail-fast cert/key checks when enabled
- IP allowlist middleware with configurable proxy-header trust
- Security headers: HSTS, X-Frame-Options, X-Content-Type-Options, Referrer-Policy

## Operational Notes
- Use strong non-default secrets for JWT and encryption keys
- Restrict allowed origins and allowed IP ranges per deployment environment
- Keep exported files in controlled local directories
