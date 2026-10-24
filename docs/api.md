# API Overview

The JSON API is available under `/api/v1`. You can view the interactive documentation at [/swagger/index.html](/swagger/index.html).

## Authentication

All API endpoints require authentication via the `Authorization: Bearer <token>` header.

## Endpoints

| Endpoint | Required Role |
|----------|---------------|
| POST /auth/register | public |
| POST /auth/login | public |
| POST /auth/refresh | public |
| POST /auth/logout | public |
| GET /me | any authenticated user |
| GET /me/loans | any authenticated user |
| GET /catalog/books | any authenticated user |
| GET /catalog/authors | any authenticated user |
| GET /catalog/categories | any authenticated user |
| POST /books | librarian and admin |
| PUT /books/:id | librarian and admin |
| POST /books/:id/copies | librarian and admin |
| POST /authors | librarian and admin |
| POST /categories | librarian and admin |
| POST /loans | librarian and admin |
| POST /loans/:id/return | librarian and admin |
| GET /loans | librarian and admin |
| DELETE /books/:id | admin only |
| GET /users | admin only |
| PUT /users/:id/role | admin only |

## Roles

- **reader**: can view catalog, view own loans
- **librarian**: can manage books, authors, categories, loans
- **admin**: can manage users and all other resources
