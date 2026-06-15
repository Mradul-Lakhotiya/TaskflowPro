# Rival Task Manager

A premium, full-stack task management application built as part of the Rival assessment.

## Tech Stack

- **Frontend**: Next.js 15 (App Router), React 19, Tailwind CSS v4, Framer Motion, Zustand
- **Backend**: Go 1.22+, `chi` router, `pgx` (PostgreSQL driver), JWT Auth
- **Database**: PostgreSQL 16+
- **DevOps**: Docker, Docker Compose, GitHub Actions

## Features Implemented

- **Core Task API**: Full CRUD operations for tasks with PostgreSQL persistence.
- **Authentication**: Secure JWT-based login/register with bcrypt password hashing.
- **Frontend Dashboard**: Beautiful, responsive UI with glassmorphism and smooth micro-animations.
- **Advanced Filtering**: Status filters, search by title, and multi-field sorting (due date, priority, creation date) working seamlessly together with pagination.
- **Optimistic UI**: Instant state updates on the frontend when toggling task completion, with automatic rollback on failure.
- **Dark Mode**: Native system-preference dark mode with custom Tailwind CSS variables.
- **Role-Based Access**: Built-in `admin` role support (can view all tasks).
- **Task Attachments**: Upload multiple images or document files directly to a task.
- **Activity Log**: View a complete historical timeline of changes made to each task.
- **CI Pipeline**: Automated testing and build verification via GitHub Actions.

## Local Setup Instructions

There are two ways to run this project: using Docker (recommended) or natively.

### 1. Dockerized Setup (One-Command Setup)
The fastest way to get the entire stack (Database, Backend, and Frontend) running:
```bash
docker compose up --build -d
```
*This spins up PostgreSQL on port `5432` (automatically running `schema.sql`), the Go backend on `8080`, and the Next.js frontend on `3000`.*

### 2. Native Setup
If you prefer to run things natively:

**Database Setup**
Ensure PostgreSQL is running locally, and execute `backend/schema.sql` to initialize the tables.

**Backend Setup**
```bash
cd backend
go mod download
# Ensure .env is created at the root level (copy from .env.example)
go run main.go
```

**Frontend Setup**
```bash
cd frontend
npm install
npm run dev
```

## Creating an Admin User

By default, all new signups are assigned the standard `user` role to prevent Mass Assignment vulnerabilities. However, you can bootstrap an `admin` account by passing a secret authorization header to the registration endpoint.

1. Ensure your `.env` file has an `ADMIN_SECRET` configured.
2. Run the following command in your terminal (replacing the URL and credentials as needed):

```bash
curl -X POST http://localhost:8080/api/auth/register \
-H "Content-Type: application/json" \
-H "X-Admin-Secret: super_secret_bootstrap_key_123" \
-d '{"email":"admin@rival.io", "password":"password123"}'
```

## Architecture, Trade-offs & Shortcomings

1. **Go Standard Library vs Heavy Frameworks**: I chose `go-chi/chi` with the standard `net/http` over heavy frameworks like Gin or Fiber. It's idiomatic, highly performant, and keeps the binary small while providing excellent routing capabilities.
2. **Pure SQL vs ORM**: I used raw SQL with `pgx` instead of GORM. For a task manager, writing raw SQL ensures maximum performance and complete control over complex queries (like dynamic filtering and sorting).
3. **Tailwind CSS v4 + Framer Motion**: Instead of relying on a bulky UI library like Material UI, I built custom components using pure Tailwind and Framer Motion. This guarantees a truly custom, premium aesthetic without the generic "bootstrap" feel.
4. **Relational Attachments vs JSONB**: I deliberately chose a separate `task_attachments` SQL table instead of a JSON array to ensure strict data integrity and atomic deletion of specific files without race conditions.
5. **Real-time updates (Shortcoming)**: While attempting to implement Server-Sent Events (SSE) for live task updates, the connection stability proved challenging and it is currently not fully working. This is a known shortcoming that I would address with more time by transitioning to robust WebSockets or refining the SSE stream management.

## Testing

The backend includes automated tests.
```bash
cd backend
go test -v ./...
```
A GitHub Actions CI pipeline is configured to run these tests automatically on push.
