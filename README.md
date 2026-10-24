# LibraTrack

LibraTrack is a library accounting system designed for schools and colleges. It manages book catalogs, loans, fines, and user roles with features tailored for different user types.

## Features by Role

### Reader
- Browse the catalog
- Check book availability
- View own loans and fines

### Librarian
- Manage books, authors, categories, and copies
- Issue and return books
- Generate reports

### Admin
- All librarian features
- Manage user roles and permissions

### Not Implemented
- Reservations queue
- Notifications

## Tech Stack

- Go 1.25+
- Gin web framework
- PostgreSQL database
- GORM ORM
- golang-migrate for database migrations
- JWT with refresh-token rotation
- bcrypt for password hashing
- html/template for rendering
- Docker for containerization

## Quick Start

## Database Seeding

To populate the database with initial seed data (such as default admin, librarian users, and initial catalog items), run the following command in your terminal:

```bash
docker compose run --rm api /app/seed
```