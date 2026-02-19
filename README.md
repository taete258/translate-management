# Translation Management System

A full-stack translation management system built with:

- **Frontend**: Svelte 5 + TypeScript + Tailwind CSS 4 (SvelteKit)
- **Backend**: Go + Fiber v2
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Containerization**: Docker Compose

## Quick Start

```bash
# Clone and start
docker compose up --build -d

# Access
# Frontend: http://localhost:5173
# Backend API: http://localhost:3000
```

## Features

- 🌐 Translation key-value management with nested key support
- 📊 Spreadsheet-style translation grid editor
- 🔑 API key management for external app integration
- 📦 Export as JSON or MessagePack format
- ⚡ Redis caching with manual cache invalidation
- 🗜️ Gzip compression for all API responses
- 🔐 User authentication with JWT
- 📈 Translation progress tracking per language

## API Endpoints

### Authentication

- `POST /api/auth/register` — Register new account
- `POST /api/auth/login` — Login
- `POST /api/auth/logout` — Logout
- `GET /api/auth/me` — Get current user info

### Projects

- `GET /api/projects` — List all projects
- `POST /api/projects` — Create a new project
- `GET /api/projects/:id` — Get project details
- `PUT /api/projects/:id` — Update project
- `DELETE /api/projects/:id` — Delete project
- `GET /api/projects/:id/stats` — Get project translation statistics
- `GET /api/projects/:id/members` — List project members

### Languages

- `GET /api/projects/:id/languages` — List project languages
- `POST /api/projects/:id/languages` — Add a new language
- `PUT /api/projects/:id/languages/:langId` — Update language settings
- `DELETE /api/projects/:id/languages/:langId` — Remove a language

### Translation Keys

- `GET /api/projects/:id/keys` — List all keys in a project
- `POST /api/projects/:id/keys` — Create a new translation key
- `PUT /api/projects/:id/keys/:keyId` — Update a translation key
- `DELETE /api/projects/:id/keys/:keyId` — Delete a translation key

### Translations

- `GET /api/projects/:id/translations` — Get all translations for a project
- `PUT /api/projects/:id/translations` — Batch update translations

### Environments

- `GET /api/projects/:id/environments` — List project environments
- `POST /api/projects/:id/environments` — Create a new environment
- `PUT /api/projects/:id/environments/:envId` — Update an environment
- `DELETE /api/projects/:id/environments/:envId` — Delete an environment

### API Keys

- `GET /api/projects/:id/api-keys` — List API keys for a project
- `POST /api/projects/:id/api-keys` — Create a new API key
- `DELETE /api/projects/:id/api-keys/:keyId` — Revoke an API key

### Invitations

- `POST /api/projects/:id/invitations` — Invite a user to a project
- `GET /api/invitations` — List current user's invitations
- `POST /api/invitations/:id/respond` — Accept or reject an invitation

### Cache & Import

- `POST /api/projects/:id/cache/invalidate` — Manually invalidate project cache
- `POST /api/projects/:id/cache/rebuild` — Rebuild project cache
- `GET /api/projects/:id/cache/status` — Get project cache status
- `POST /api/projects/:id/import` — Import translations from JSON

### Export (External API)

- `GET /api/export/:slug/:langCode?format=json|msgpack` — External export using API Key
- `GET /api/export/:slug/:langCode/version` — Get current version hash
- `GET /api/projects/:id/export/:langCode` — Direct export for frontend (JWT protected)

## Environment Variables

Copy `.env.example` to `.env` and configure:

```env
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=translate_management
REDIS_HOST=redis
REDIS_PORT=6379
JWT_SECRET=your-secret-key
PORT=3000
PUBLIC_API_URL=http://localhost:3000
```
