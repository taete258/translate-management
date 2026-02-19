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
- `GET /api/auth/me` — Get current user

### Projects

- `GET/POST /api/projects` — List/Create projects
- `GET/PUT/DELETE /api/projects/:id` — Get/Update/Delete project

### Languages

- `GET/POST /api/projects/:id/languages` — List/Add languages
- `PUT/DELETE /api/projects/:id/languages/:langId` — Update/Remove language

### Translation Keys

- `GET/POST /api/projects/:id/keys` — List/Create keys
- `PUT/DELETE /api/projects/:id/keys/:keyId` — Update/Delete key

### Translations

- `GET /api/projects/:id/translations` — Get all translations
- `PUT /api/projects/:id/translations` — Batch update translations

### API Keys

- `GET/POST /api/projects/:id/api-keys` — List/Create API keys
- `DELETE /api/projects/:id/api-keys/:keyId` — Revoke API key

### Cache

- `POST /api/projects/:id/cache/invalidate` — Force invalidate cache

### Export (API Key Required)

- `GET /api/export/:slug/:langCode?format=json|msgpack` — Export translations
- `GET /api/export/:slug/:langCode/version?format=json|msgpack` — Get export version hash

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
