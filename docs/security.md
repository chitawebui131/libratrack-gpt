# Security

## Implemented

- bcrypt password hashing
- short-lived JWT access tokens
- rotating refresh tokens with reuse detection
- HttpOnly SameSite=Lax cookies for the web UI (Secure via COOKIE_SECURE)
- CSRF protection with a double-submit cookie
- Content-Security-Policy and other security headers
- RBAC (reader, librarian, admin)
- role loaded from the database on web requests
- cookies are not accepted by /api/v1

## Before production

- HTTPS and COOKIE_SECURE=true
- rate limiting
- backups
- secrets management
- JWT_SECRET rotation
