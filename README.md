# Unlupa API (Hifzhun Backend)

REST API backend untuk platform **Unlupa (Hifzhun)** menggunakan Go (Golang), Fiber v2, GORM, PostgreSQL, dan Redis.

---

## 🛠 Tech Stack

- **Language**: Go 1.24+
- **Web Framework**: Fiber v2 (`github.com/gofiber/fiber/v2`)
- **Database & ORM**: PostgreSQL, GORM (`gorm.io/gorm`)
- **Caching**: Redis
- **Auth**: JWT (JSON Web Token), bcrypt
- **Algorithm Engine**: Free Spaced Repetition Scheduler (FSRS)
- **Object Storage**: Supabase Storage (S3-compatible)

---

## 📂 Struktur Direktori

```
unlupa-api/
├── api/
│   ├── handlers/       # Controller / HTTP Route Handlers
│   ├── middlewares/    # Auth, Logger, CORS, Rate Limit
│   └── routes/         # Router Group Registrations
├── pkg/
│   ├── config/         # Konfigurasi Database, Redis, App
│   ├── entities/       # GORM Model Database Entities
│   ├── fsrs/           # FSRS Algorithm Implementation
│   ├── repositories/   # Database Query & Persistence Layer
│   ├── services/       # Business Logic Layer
│   └── utils/          # Standard Response, Hashing, Validation
├── docs/               # Swagger / OpenAPI Documentation
├── uploads/            # Temporary File Uploads
├── main.go             # Application Entry Point
└── go.mod              # Go Module Definitions
```

---

## 🚀 Cara Menjalankan

### 1. Prasyarat
- Go 1.24 atau lebih baru terpasang di sistem.
- PostgreSQL & Redis telah berjalan.

### 2. Konfigurasi Environment (`.env`)
Salin atau buat file `.env` di folder root `unlupa-api/`:

```env
# Application
APP_PORT=3000
APP_MODE=dev

# PostgreSQL Database
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=hifzhun_db
DB_SSLMODE=disable
DB_TIMEZONE=Asia/Jakarta

# JWT Security
JWT_SECRET=your_secret_key_here
JWT_EXPIRE_HOURS=24

# Redis Cache
REDIS_URL=localhost:6379

# Supabase Storage (Opsional)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
SUPABASE_BUCKET=unlupa-storage
```

### 3. Install Dependensi
```bash
go mod download
```

### 4. Jalankan Server
```bash
go run main.go
```
Server akan aktif di: `http://localhost:3000`

---

## 📡 Ringkasan API Endpoints

Base URL: `/api/v1`

### 1. Autentikasi (`/auth`)
| Method | Endpoint | Deskripsi |
|---|---|---|
| `POST` | `/auth/register` | Mendaftarkan akun pengguna baru (`student` / `teacher`) |
| `POST` | `/auth/login` | Masuk dan mendapatkan JWT Token |
| `GET` | `/auth/me` | Mendapatkan data profil pengguna saat ini |
| `PUT` | `/auth/admin/approve/:id` | Menyetujui akun guru (Khusus Admin) |

### 2. Katalog & Halaman Al-Qur'an (`/quran`)
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/quran/juzs?user_id=` | Mengambil 30 Juz beserta rekap status aktif, mapan, dan due today |
| `GET` | `/quran/juzs/:juzNumber/pages?user_id=` | Mengambil detail halaman Mushaf Madani per Juz |
| `POST` | `/quran/pages/:mushafPage/activate` | Mengaktifkan halaman Al-Qur'an untuk mulai dihafal |
| `POST` | `/quran/pages/review` | Mengirimkan penilaian murajaah FSRS (Rating 1-4) |
| `GET` | `/quran/pages?user_id=` | Rekap kemajuan seluruh 604 halaman Mushaf |
| `GET` | `/quran/juz30` | Checklist khusus progres Juz 30 (Surah 78-114) |

### 3. Ruang Kelas (`/classes`)
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/classes` | Mengambil daftar kelas yang dikelola pengajar |
| `GET` | `/classes/joined` | Mengambil daftar kelas yang diikuti santri |
| `POST` | `/classes` | Membuat ruang kelas baru |
| `POST` | `/classes/join` | Bergabung ke kelas menggunakan kode kelas (6 karakter) |
| `GET` | `/classes/:id/members` | Mengambil daftar santri dalam kelas beserta progres |
| `POST` | `/classes/:id/leave` | Santri keluar dari kelas (*Leave Class*) |
| `PUT` | `/classes/:id` | Memperbarui nama, deskripsi, atau cover kelas |
| `DELETE` | `/classes/:id` | Menghapus kelas (Hanya Guru Pemilik) |

---

## 🧪 Testing

Jalankan test suite menggunakan:
```bash
go test ./...
```
