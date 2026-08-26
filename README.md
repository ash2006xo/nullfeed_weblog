# nullfeed weblog

A small weblog built with **Go, Echo, PostgreSQL and JavaScript**. The project keeps a deliberately dark, minimal visual style while focusing on writing, private sharing and conversations.

## Features

- Username/password authentication with bcrypt + HTTP-only JWT cookies
- Public and private blog posts
- Private sharing by username
- Owner-only post deletion
- Optional image uploads (JPEG, PNG, GIF, WebP, 5 MB max)
- Comments for users who can view a post
- Responsive, mobile-friendly frontend
- Search and public/private feed filters
- PostgreSQL schema initialization on startup
- Docker / Render deployment configuration

## Stack

- Go
- Echo
- PostgreSQL
- HTML/CSS/vanilla JavaScript
- JWT + bcrypt

## Run locally

1. Start PostgreSQL with Docker Compose:

```bash
docker compose up -d
```

2. Copy the environment file:

```bash
cp .env.example .env
```

3. Start the server:

```bash
go run ./cmd/server
```

4. Open `http://localhost:8080`.

The application creates the required database tables automatically when it starts.

## Environment

Required:

```text
DATABASE_URL=postgres://nullfeed:devpassword@localhost:5432/nullfeed?sslmode=disable
JWT_SECRET=replace-with-a-long-random-secret
PORT=8080
```

## Routes

### Pages

- `/` — feed
- `/login.html` — login
- `/signup.html` — signup
- `/create.html` — create a post
- `/weblog/:id` — post detail

### API

- `POST /api/signup`
- `POST /api/login`
- `POST /api/logout`
- `GET /api/me`
- `GET /api/boards`
- `GET /api/boards/:id`
- `POST /api/boards`
- `DELETE /api/boards/:id`
- `GET /api/boards/:id/shares`
- `PUT /api/boards/:id/shares`
- `POST /api/uploads/image`
- `GET /api/boards/:id/comments`
- `POST /api/boards/:id/comments`

Repo : 
https://github.com/ash2006xo/nullfeed_weblog

Live application:
https://nullfeed.onrender.com
