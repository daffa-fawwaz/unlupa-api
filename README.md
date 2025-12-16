# Hifzhun API

REST API untuk aplikasi Hifzhun menggunakan Go, Fiber, dan PostgreSQL.

## Tech Stack

- Go 1.24+
- Fiber v2 (Web Framework)
- GORM (ORM)
- PostgreSQL
- JWT Authentication
- bcrypt (Password Hashing)

## Struktur Project

```
hifzhun-api/
├── api/
│   ├── handlers/    # HTTP handlers
│   └── routes/      # Route definitions
├── pkg/
│   ├── config/      # Database config
│   ├── entities/    # GORM models
│   ├── repositories/# Data access layer
│   ├── services/    # Business services
│   ├── usecases/    # Business logic
│   └── utils/       # Helper functions
├── main.go
└── go.mod
```

## Setup

### 1. Clone repository

```bash
git clone <repository-url>
cd hifzhun-api
```

### 2. Buat file `.env`

```env
# App
APP_PORT=3000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=hifzhun
DB_SSLMODE=disable
DB_TIMEZONE=Asia/Jakarta

# JWT
JWT_SECRET=your-secret-key
```

### 3. Install dependencies

```bash
go mod download
```

### 4. Jalankan server

```bash
go run main.go
```

Server akan berjalan di `http://localhost:3000`

## API Endpoints

Base URL: `/api/v1`

### Auth

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/auth/register` | Register user baru |
| POST | `/auth/login` | Login user |
| PUT | `/auth/admin/approve/:id` | Approve teacher (admin only) |

### Register

```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "email": "john@example.com",
    "password": "password123",
    "role": "student",
    "full_name": "John Doe"
  }'
```

**Role yang tersedia:**
- `student` - langsung aktif
- `teacher` - perlu approval admin

### Login

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "login success",
  "data": {
    "id": "uuid",
    "email": "john@example.com",
    "role": "student",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### Approve Teacher (Admin)

```bash
curl -X PUT http://localhost:3000/api/v1/auth/admin/approve/{user_id}
```

## Database

Project menggunakan GORM AutoMigrate. Tabel akan dibuat otomatis saat server dijalankan:

- `users`
- `kitabs`
- `classes`
- `class_members`
- `cards`
- `card_states`
