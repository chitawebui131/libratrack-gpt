# Configuration

This application uses environment variables for configuration. All variables are optional and will use default values if not set.

| Variable | Default | Description |
|----------|---------|-------------|
| JWT_SECRET | `change-me-to-a-random-string-of-at-least-32-characters` | Secret key for signing JSON Web Tokens. Must be at least 32 characters long. |
| JWT_TTL_MINUTES | `60` | Time-to-live for JSON Web Tokens in minutes. |
| ADMIN_EMAIL | `admin@libratrack.local` | Email address for the admin user. |
| ADMIN_PASSWORD | `change-me-admin-password` | Password for the admin user. This value is required and must be changed. |
| SEED_DEMO_USERS | `false` | Whether to seed demo users during initialization. |
| DEMO_USERS_PASSWORD | `change-me-demo-password` | Password for demo users. This value is required and must be changed. |
| LOAN_PERIOD_DAYS | `14` | Default loan period in days. |
| FINE_PER_DAY | `2` | Fine amount per day for overdue books. |
| FINE_MAX | `100` | Maximum fine amount for a single book. |
| MAX_ACTIVE_LOANS | `5` | Maximum number of active loans per user. |
| COOKIE_SECURE | `false` | Whether to set secure flag on cookies. |
| APP_TIMEZONE | `Europe/Kyiv` | Application timezone. |
| APP_ENV | `development` | Application environment (development, staging, production). |
