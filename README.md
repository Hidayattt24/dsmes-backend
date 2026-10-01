# DSMES Aceh Backend — Developer & Architecture Guide

Pusat backend REST API untuk platform **DSMES Aceh** (*Diabetes Self-Management Education and Support*). Backend ini menjadi pusat aturan bisnis, klasifikasi medis, autentikasi multi-peran, kalkulasi nutrisi, dan penyimpanan data bagi dua platform klien: **Aplikasi Mobile Pasien (Flutter)** dan **Portal Web Admin/Staff (Next.js)**.

---

## Daftar Isi
1. [Arsitektur Integrasi Multi-Platform (Mobile & Web)](#arsitektur-integrasi-multi-platform-mobile--web)
2. [Tech Stack & Pustaka Utama](#tech-stack--pustaka-utama)
3. [Struktur Proyek & Clean Modular Pattern (`internal/`)](#struktur-proyek--clean-modular-pattern-internal)
4. [Standar Penamaan Variabel & Kode (Coding Standards)](#standar-penamaan-variabel--kode-coding-standards)
5. [Logika Medis & Single Source of Truth](#logika-medis--single-source-of-truth)
6. [Arsitektur Database, GORM & Migrasi](#arsitektur-database-gorm--migrasi)
7. [Konfigurasi Environment (`.env`)](#konfigurasi-environment-env)
8. [Konfigurasi Docker & Containerisasi](#konfigurasi-docker--containerisasi)
9. [Konfigurasi CI/CD Pipeline & Deployment VPS](#konfigurasi-cicd-pipeline--deployment-vps)
10. [Instalasi Lokal & Skrip Verifikasi](#instalasi-lokal--skrip-verifikasi)

---

## Arsitektur Integrasi Multi-Platform (Mobile & Web)

Backend melayani dua klien dengan kebutuhan berbeda melalui satu basis kode (Unified REST API):

```text
 ┌─────────────────────────┐               ┌─────────────────────────┐
 │   Aplikasi Mobile Pasien│               │   Portal Web Admin/Staff│
 │         (Flutter)       │               │        (Next.js)        │
 └────────────┬────────────┘               └────────────┬────────────┘
              │                                         │
              │  /api/v1/patient/*                      │  /api/v1/admin/*
              │  /api/v1/routines/*                     │  /api/v1/staff/*
              │                                         │
              └───────────────────┬─────────────────────┘
                                  │
                      ┌───────────▼───────────┐
                      │   DSMES Aceh Backend  │
                      │    (Go Fiber v3)      │
                      └───────────┬───────────┘
                                  │
                      ┌───────────▼───────────┐
                      │  PostgreSQL Database  │
                      └───────────────────────┘
```

### Multi-Role Access Control (RBAC)
* **Role `patient` (Mobile Flutter):** Digunakan oleh penderita diabetes melitus. Mengakses pencatatan glukosa mandiri, log makanan dan kalori, aktivitas fisik, jadwal pengingat (reminder/routine), kuis, dan pengisian survei penelitian.
* **Role `staff` (Web Portal Puskesmas):** Digunakan oleh dokter/perawat di Puskesmas. Memantau tren gula darah populasi pasien terdaftar di wilayah fasilitasnya, melihat kepatuhan minum obat, dan menganalisis respons kuesioner pasien.
* **Role `admin` (Web Portal Dinas/Superadmin):** Memiliki hak penuh atas konfigurasi sistem, data faskes, akun staff tenaga medis, data seluruh pasien, materi edukasi, bank soal assessment, serta instrumen survei penelitian (SUS & Kepuasan).

---

## Tech Stack & Pustaka Utama

* **Language:** Go `1.26.4`
* **HTTP Framework:** Go Fiber `v3.3.0` (Fast HTTP router, zero memory allocation)
* **ORM:** GORM `v1.31.2` dengan PostgreSQL driver (`pgx/v5`)
* **Database:** PostgreSQL `15+`
* **Authentication:** JWT `golang-jwt/jwt/v5` + Password hashing `bcrypt`
* **Validation:** `go-playground/validator/v10`
* **Configuration:** Viper (Twelve-factor app configuration)
* **Logger:** Uber Zap (High-performance structured JSON logging)
* **API Documentation:** Swaggo / Swagger UI OpenAPI spec
* **Email Service:** Resend API (Pengiriman token OTP & reset password)

---

## Struktur Proyek & Clean Modular Pattern (`internal/`)

Backend menggunakan pendekatan **Domain-Driven Modular Architecture**:

```text
dsmes-backend/
├── cmd/
│   ├── api/
│   │   ├── main.go              # Entry point bootstrap server
│   │   └── routes.go            # Registrasi rute HTTP Fiber & middleware group
│   ├── migrate/
│   │   └── main.go              # CLI Runner migrasi database terstruktur
│   └── seed/
│       └── main.go              # CLI Seeder data awal (admin, faskes, instrumen)
│
├── internal/
│   ├── domain/                  # Pure Core Domain Entities & Contract Enums
│   │   ├── patient.go           # Struct Patient, Gender, AccountStatus
│   │   ├── blood_sugar.go       # Struct BloodSugarLog, GlucoseSeverity
│   │   ├── nutrition.go         # Struct MealLog, CalorieRecommendation
│   │   ├── survey.go            # Struct Survey, Question, SurveyResponse
│   │   ├── reminder.go          # Struct Routine, ReminderLog
│   │   └── staff.go             # Struct Staff, StaffRole
│   │
│   ├── modules/                 # Feature Modules (Vertical Slicing)
│   │   ├── auth/                # Login, Register, Refresh Token, OTP
│   │   ├── patient/             # Profil, data medis, sosiodemografi, compliance
│   │   ├── survey/              # Survey SUS & Kepuasan Pengguna
│   │   ├── quiz/                # Bank soal, kuis pre/post-test
│   │   ├── education/           # Artikel edukasi, video, tracking progress
│   │   ├── nutrition/           # Database makanan, tracking kalori
│   │   ├── reminder/            # Jadwal pengingat rutin harian
│   │   ├── routine/             # Onboarding rutinitas pasien
│   │   └── staff/               # Manajemen akun dan hak akses staff
│   │
│   ├── bootstrap/               # Inisialisasi Database, Logger, dan Fiber
│   └── middleware/              # Auth JWT, RBAC Role Guard, CORS, Panic Recovery
│
├── migrations/                  # File migrasi SQL mentah (000001_...up.sql & down.sql)
├── docs/                        # Generated Swagger JSON/YAML
└── Dockerfile                   # Multi-stage production container build
```

### Anatomi Setiap Modul (`internal/modules/<fitur>/`)
Setiap modul fitur terisolasi dengan struktur internal seragam:
1. `interfaces.go`: Kontrak interface untuk `Repository` dan `Service`.
2. `handler.go`: Layer HTTP. Mengurai request body, validasi payload, dan mengembalikan response JSON.
3. `service.go`: Layer logika bisnis. Tempat kalkulasi medis, integrasi antar-domain, dan orchestration.
4. `repository.go`: Layer akses data. Eksekusi query database GORM ke PostgreSQL.
5. `dto.go`: Data Transfer Object untuk binding request dan response serializer.

---

## Standar Penamaan Variabel & Kode (Coding Standards)

Untuk menjaga konsistensi pada codebase Go:

| Komponen | Standar | Contoh | Keterangan |
| :--- | :--- | :--- | :--- |
| **Exported Identifiers** | `PascalCase` | `PatientService`, `ListPatients` | Dapat diakses dari luar package |
| **Unexported Identifiers**| `camelCase` | `patientRepo`, `calculateBMI` | Hanya dipakai internal package |
| **Struct Fields (Go)** | `PascalCase` | `FullName`, `SmokingStatus` | Field pada model domain / DTO |
| **JSON DTO Tag** | `snake_case` | `json:"smoking_status"` | Kunci JSON pada body request/response |
| **Database Column** | `snake_case` | `date_of_birth`, `health_facility` | Nama kolom di tabel PostgreSQL |
| **URL Endpoints** | `kebab-case` | `/api/v1/patient/meal-logs` | Standar rute REST API |
| **Query Parameters** | `snake_case` | `?compliance_min=70&page=1` | Standar parameter pencarian URL |

---

## Logika Medis & Single Source of Truth

**Penting:** Seluruh klien (Web Portal dan Mobile Flutter) **dilarang** mengkalkulasi status klinis atau skor kepatuhan secara independen di sisi klien. Backend menjadi **Single Source of Truth** untuk:

### 1. Klasifikasi Gula Darah (`domain.CalculateGlucoseStatus`)
Glukosa darah diklasifikasikan berdasarkan waktu pengukuran (Puasa, 2 Jam PP, Sewaktu, Sebelum Tidur) ke dalam kategori:
* `normal`: Rentang target aman
* `hipoglikemia` & `severe_hypoglycemia`: Gula darah di bawah batas aman
* `prediabetes` / `tinggi`: Gula darah di atas batas optimal
* `hiperglikemia` & `severe_hyperglycemia`: Gula darah tinggi membutuhkan intervensi

### 2. Kepatuhan Pasien (`complianceFromAggregates`)
Skor kepatuhan dihitung harian dari agregat pencatatan:
* `≥ 70%`: **Patuh** (Tercapai target pemantauan harian)
* `40% - 69%`: **Kurang Patuh**
* `< 40%`: **Tidak Patuh** (Perlu perhatian tenaga medis Puskesmas)

### 3. Kebutuhan Kalori Harian (Rumus DSMES)
Dihitung otomatis berdasarkan jenis kelamin, tinggi badan, berat badan aktual, usia, dan tingkat aktivitas fisik pasien.

---

## Arsitektur Database, GORM & Migrasi

### 1. Migrasi Terstruktur
Migrasi dikelola melalui runner terpusat `cmd/migrate`. Versi migrasi dicatat pada tabel `dsmes_migrations`.
```bash
# Menjalankan migrasi database
go run ./cmd/migrate
```

### 2. Aturan Query GORM (Pencegahan SQL Ambiguity)
Karena tabel-tabel di DSMES menggunakan soft-delete (`deleted_at`), setiap query yang menggabungkan beberapa tabel (`Joins` atau `Preload`) **wajib mencantumkan nama tabel secara eksplisit**.

Contoh pola yang benar:
```go
// BENAR: Menggunakan kualifikasi nama tabel secara eksplisit
db.Preload("Responses", func(db *gorm.DB) *gorm.DB {
    return db.Where("survey_responses.deleted_at IS NULL").
        Joins("JOIN patients p ON p.id = survey_responses.patient_id AND p.deleted_at IS NULL").
        Where("p.health_facility = ?", facilityName)
})
```
*Hindari menulis `Where("deleted_at IS NULL")` tanpa nama tabel pada query berelasi karena PostgreSQL akan mengembalikan `ERROR: column reference "deleted_at" is ambiguous` (HTTP 500).*

---

## Konfigurasi Environment (`.env`)

Buat file `.env` di folder root `dsmes-backend`:

```env
# Server
APP_NAME=dsmes-backend
APP_ENV=development
APP_PORT=8080
APP_BASE_URL=http://localhost:8080
APP_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
APP_TIMEZONE=Asia/Jakarta

# Database PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=dsmes_db
DB_USER=dsmes_user
DB_PASSWORD=secret_password
DB_SSLMODE=disable

# JWT Authentication
JWT_SECRET=super_secret_jwt_key_min_32_characters
JWT_ACCESS_TOKEN_TTL=15m
JWT_REFRESH_TOKEN_TTL=168h
JWT_ISSUER=dsmes-backend

# Resend Mail (OTP & Password Reset)
RESEND_API_KEY=re_xxxxxxxxxxxx
RESEND_FROM_EMAIL=no-reply@dsmes-aceh.id

# Dokumentasi Swagger
SWAGGER_ENABLED=true
SWAGGER_HOST=localhost:8080
```

---

## Konfigurasi Docker & Containerisasi

### Multi-stage Dockerfile (`Dockerfile`)
* **Stage 1 (Builder):** Menggunakan `golang:1.24-alpine`, meng-compile binary statis `api` dan `migrate`, serta menghasilkan Swagger docs.
* **Stage 2 (Runtime):** Menggunakan `alpine:latest` dengan user non-root `appuser`.
* **Entrypoint (`entrypoint.sh`):** Secara otomatis mengeksekusi migrasi database (`/app/migrate`) sebelum menyalakan API server (`/app/api`).

### Menjalankan dengan Docker Compose
```bash
# Menjalankan PostgreSQL lokal
docker compose up -d postgres

# Menjalankan seluruh stack backend
docker compose --profile app up -d
```

---

## Konfigurasi CI/CD Pipeline & Deployment VPS

File workflow CI/CD berada di `.github/workflows/ci.yml`. Pipeline berjalan otomatis pada event `push` dan `pull_request`:

### 1. Continuous Integration (CI)
* Memeriksa static analysis kode (`go vet ./...`).
* Menjalankan unit & integration test dengan race detection (`go test -race -count=1 ./...`).
* Memeriksa kelulusan kompilasi binary `cmd/api` dan `cmd/migrate`.

### 2. Continuous Deployment (CD ke VPS)
Ketika commit di-merge ke branch `main`:
1. GitHub Actions meng-compile Docker Image dan mempublikasikannya ke GitHub Container Registry (`ghcr.io`).
2. Melakukan koneksi SSH ke VPS produksi.
3. Menarik (*pull*) image terbaru dan menjalankan zero-downtime update via `docker-compose.production.yml`.
4. Menguji health check endpoint (`GET /api/health`). Jika gagal setelah 30x percobaan, pipeline secara otomatis melakukan rollback ke image stabil sebelumnya.

---

## Instalasi Lokal & Skrip Verifikasi

### Menjalankan Backend Lokal
```bash
# 1. Download module dependencies
go mod download

# 2. Jalankan migrasi database
go run ./cmd/migrate

# 3. Jalankan REST API server
go run ./cmd/api
```
Layanan aktif pada:
* **API Root:** `http://localhost:8080`
* **Health Check:** `http://localhost:8080/api/health`
* **Swagger UI:** `http://localhost:8080/swagger/index.html`

### Menjalankan Uji Mutu Kode (Self-Verification)
```bash
# 1. Format & vet
go vet ./...

# 2. Jalankan test suite
go test ./...

# 3. Kompilasi binary API
go build -o /dev/null ./cmd/api
```
