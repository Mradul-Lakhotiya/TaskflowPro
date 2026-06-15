# Rival Task Manager

A premium, full-stack task management application built as part of the Rival assessment.

## 🚀 Tech Stack

- **Frontend**: Next.js 15 (App Router), React 19, Tailwind CSS v4, Framer Motion, Zustand
- **Backend**: Go 1.22+, `chi` router, `pgx` (PostgreSQL driver), JWT Auth
- **Database**: PostgreSQL 16+
- **DevOps**: Docker, Docker Compose, GitHub Actions

## ✨ Features Implemented

- **Core Task API**: Full CRUD operations for tasks with PostgreSQL persistence.
- **Authentication**: Secure JWT-based login/register with bcrypt password hashing.
- **Frontend Dashboard**: Beautiful, responsive UI with glassmorphism and smooth micro-animations.
- **Advanced Filtering**: Status filters, search by title, and multi-field sorting (due date, priority, creation date) working seamlessly together with pagination.
- **Optimistic UI**: Instant state updates on the frontend when toggling task completion, with automatic rollback on failure.
- **Dark Mode**: Native system-preference dark mode with custom Tailwind CSS variables.
- **Role-Based Access**: Built-in `admin` role support (can view all tasks, though UI defaults to viewing own tasks).

## 🛠️ Local Setup Instructions

There are two ways to run this project: using Docker (recommended for Codespaces) or natively.

### Prerequisites
- Node.js v20+
- Go v1.22+
- PostgreSQL (if not using Docker)

### 1. Database Setup (Docker - Recommended)
The fastest way to get the database running, especially in GitHub Codespaces:
```bash
docker-compose up -d
```
*This starts a PostgreSQL container on port 5432 and automatically applies the schema from `backend/schema.sql`.*

### 2. Backend Setup
```bash
cd backend
go mod download
# Make sure .env is created at the root level (copy from .env.example)
go run main.go
```
*The backend will run on `http://localhost:8080`.*

### 3. Frontend Setup
```bash
cd frontend
npm install
npm run dev
```
*The frontend will run on `http://localhost:3000`.*

## 🔐 Creating an Admin User

By default, all new signups are assigned the standard `user` role to prevent Mass Assignment vulnerabilities. However, you can bootstrap an `admin` account by passing a secret authorization header to the registration endpoint.

1. Ensure your `.env` file has an `ADMIN_SECRET` configured.
2. Run the following command in your terminal (replacing the URL and credentials as needed):

```bash
curl -X POST http://localhost:8080/api/auth/register \
-H "Content-Type: application/json" \
-H "X-Admin-Secret: super_secret_bootstrap_key_123" \
-d '{"email":"admin@rival.io", "password":"password123"}'
```
*If the secret perfectly matches the backend config, the new account will be permanently elevated to Admin.*

## 📂 Architecture & Trade-offs

1. **Go Standard Library vs Heavy Frameworks**: I chose `go-chi/chi` with the standard `net/http` over heavy frameworks like Gin or Fiber. It's idiomatic, highly performant, and keeps the binary small while providing excellent routing capabilities.
2. **Pure SQL vs ORM**: I used raw SQL with `pgx` instead of GORM. For a task manager, writing raw SQL ensures maximum performance and complete control over complex queries (like dynamic filtering and sorting).
3. **Tailwind CSS v4 + Framer Motion**: Instead of relying on a bulky UI library like Material UI, I built custom components using pure Tailwind and Framer Motion. This guarantees a truly custom, premium aesthetic without the generic "bootstrap" feel.
4. **Optimistic UI**: Implemented on the task completion toggle. It provides a highly responsive feel. If the server request fails, the UI rolls back automatically.
5. **Bonus Features Deferred**: As discussed, WebSockets and File Attachments were deferred to focus on delivering a highly polished core experience first, keeping the architecture extensible for future integration.

## 🧪 Testing

The backend includes automated tests for the authentication utilities and middleware.
```bash
cd backend
go test -v ./...
```
A GitHub Actions CI pipeline is configured to run these tests automatically on push.
