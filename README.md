# DSMES Aceh Backend

Backend REST API untuk platform DSMES Aceh (Diabetes Self-Management Education and Support). Backend menjadi pusat aturan bisnis, autentikasi, validasi, penyimpanan data, klasifikasi medis, dan komunikasi antara aplikasi pasien, portal web, serta PostgreSQL.

## Tujuan Sistem

Backend dibangun untuk membantu pasien mengelola diabetes secara mandiri dan membantu staff/admin memantau perkembangan pasien. Seluruh client menggunakan API yang sama agar data dan aturan bisnis tetap konsisten.

## Fitur Utama

- Authentication: register, login, JWT access token, refresh token, logout, OTP, dan reset password.
- Role-based access: pasien, staff/Puskesmas, dan admin.
- Patient profile: data pribadi, sosiodemografi, kondisi kesehatan, fasilitas kesehatan, dan foto profil.
- Health records: gula darah, berat/tinggi badan, BMI, makanan, kalori, aktivitas fisik, obat, dan reminder.
- Blood sugar classification: kategori, severity, rentang referensi, rekomendasi, dan warna dihasilkan oleh classifier backend.
- Education: artikel, video, progress belajar, dan review.
- Assessment: pre-test, questionnaire, quiz, survey, scoring, dan hasil assessment.
- Monitoring: dashboard, statistik agregat, riwayat kesehatan, dan pemantauan pasien.
- AI assistant: percakapan bantuan personal untuk edukasi diabetes.
- Staff management: pengelolaan akun dan akses staff.
- Infrastructure: PostgreSQL migrations, structured logging, CORS, recovery middleware, dan Swagger UI.

## Teknologi

- Go `1.26.4`
- Go Fiber `3.3.0`
- PostgreSQL dengan GORM `1.31.2` dan pgx
- JWT `v5` dan bcrypt
- Viper untuk konfigurasi
- Validator v10
- Zap untuk logging
- Swagger untuk dokumentasi API
- Resend untuk email transaksional/OTP

## Struktur Proyek

```text
cmd/api/          Entry point REST API
cmd/migrate/      Database migration runner
cmd/seed/         Data seed
internal/domain/  Entity dan aturan domain
internal/modules/ Modul fitur backend
internal/bootstrap/Database, logger, dan aplikasi Fiber
internal/middleware/ Auth, RBAC, CORS, recovery, dan logging
migrations/       SQL migration up/down
docs/              Dokumentasi Swagger
```

## Prasyarat

- Go `1.26.4` atau lebih baru
- PostgreSQL `15` atau `16`
- Docker Compose (opsional)
- Air (opsional untuk live reload)

## Konfigurasi Environment

Salin template konfigurasi:

```powershell
Copy-Item .env.example .env
```

Key yang tersedia:

```text
APP_NAME
APP_ENV                 # development, staging, atau production
APP_PORT
APP_BASE_URL
APP_ALLOWED_ORIGINS
APP_TIMEZONE
APP_READ_TIMEOUT
APP_WRITE_TIMEOUT
APP_IDLE_TIMEOUT

DB_HOST
DB_PORT
DB_NAME
DB_USER
DB_PASSWORD
DB_SSLMODE
DB_MAX_IDLE_CONNS
DB_MAX_OPEN_CONNS
DB_CONN_MAX_LIFETIME_MINUTES

JWT_SECRET
JWT_ACCESS_TOKEN_TTL
JWT_REFRESH_TOKEN_TTL
JWT_ISSUER

LOG_LEVEL
LOG_FORMAT
SWAGGER_ENABLED
SWAGGER_HOST

RESEND_API_KEY
RESEND_FROM_EMAIL

AI_CHATBOT
AI_PROVIDER
AI_MODEL
AI_LOG_PROMPTS
```

Jangan commit `.env`, password database, JWT secret, email key, atau AI key. User database runtime tidak harus menjadi owner tabel; jalankan migration dengan database owner/superuser jika diperlukan.

## Menjalankan Backend

```powershell
go mod download
go run ./cmd/api
```

Dengan Air:

```powershell
air
```

Health check:

```text
GET http://localhost:8080/api/health
```

Swagger:

```text
http://localhost:8080/swagger/index.html
```

## Database Migration

Jalankan dari root `dsmes-backend` menggunakan user database yang memiliki hak mengubah schema:

```powershell
go run ./cmd/migrate
```

Migration baru dicatat pada tabel `dsmes_migrations`. Jangan menjalankan migration production menggunakan user aplikasi yang tidak memiliki hak owner schema.

## CI/CD dan DevOps

Workflow CI/CD berada di:

```text
.github/workflows/ci.yml
```

Pipeline berjalan pada setiap push ke branch `main` atau `develop`, serta setiap Pull Request.

### Continuous Integration (CI)

Job `test` menjalankan pemeriksaan berikut pada runner Ubuntu:

1. Checkout source code.
2. Setup Go sesuai versi yang ditentukan workflow.
3. Menjalankan `golangci-lint`.
4. Menjalankan `go vet ./...`.
5. Menjalankan `go test -race -count=1 ./...`.
6. Build binary API dan migration runner.
7. Generate Swagger documentation.

Pull Request harus melewati job CI sebelum perubahan digabungkan.

### Container Build

Setelah job CI berhasil, job `docker`:

- Membuat multi-stage Docker image menggunakan `Dockerfile`.
- Menghasilkan binary statis untuk Linux.
- Menyertakan server, migration runner, migrations, dan Swagger docs.
- Memberi tag image berdasarkan branch, Pull Request, dan commit SHA.
- Push image ke GitHub Container Registry (`ghcr.io`) untuk push non-PR.
- Menggunakan GitHub Actions cache untuk mempercepat build berikutnya.

Pull Request hanya melakukan build image tanpa push ke registry.

### Continuous Deployment (CD)

Deployment production hanya berjalan ketika push berhasil ke branch `main`.

Alurnya:

```text
Push ke main
    -> CI lint, vet, test, dan build
    -> Build Docker image
    -> Push image ke GHCR
    -> SSH ke VPS
    -> Pull image berdasarkan commit SHA
    -> docker compose up -d
    -> Health check API
    -> Rollback ke image sebelumnya jika health check gagal
```

Deployment menggunakan:

```text
docker-compose.production.yml
```

Health check production:

```text
GET http://127.0.0.1:8080/api/health
```

Pipeline menunggu health check hingga 30 percobaan. Jika service tidak sehat, container sebelumnya digunakan kembali bila image sebelumnya tersedia.

### GitHub Actions Secrets

Secret berikut dikonfigurasi pada GitHub Actions, terutama environment `production`:

```text
VPS_HOST
VPS_USER
VPS_SSH_KEY
VPS_KNOWN_HOSTS
GHCR_DEPLOY_TOKEN
```

`GITHUB_TOKEN` disediakan otomatis oleh GitHub Actions untuk push ke GHCR sesuai permission workflow. Jangan menulis nilai secret di workflow, source code, README, atau log.

### Docker Runtime

`Dockerfile` menggunakan dua stage:

```text
builder  -> compile server, migrate, dan generate Swagger
runtime  -> Alpine minimal dengan user non-root
```

Container menjalankan `/app/migrate` sebelum `/app/server` melalui `entrypoint.sh`. Migration harus idempotent dan database user yang digunakan container harus memiliki hak schema yang diperlukan. Jika runtime user tidak memiliki hak `ALTER TABLE`, gunakan migration job terpisah dengan database owner sebelum deployment aplikasi.

### Local DevOps Commands

```powershell
# Menyalakan PostgreSQL local
docker compose up -d postgres

# Menjalankan seluruh service Docker
docker compose --profile app up -d

# Melihat log service
docker compose logs -f

# Build image local
docker build -t dsmes-backend:local .

# Menjalankan quality checks yang sama secara manual
go vet ./...
go test -race -count=1 ./...
go build ./cmd/api
```

Untuk shortcut yang tersedia, gunakan `make help`. Target penting meliputi `make test`, `make lint`, `make migrate`, `make docker-up`, dan `make docker-build`.

## Testing dan Build

```powershell
go test ./...
go build -o bin/api ./cmd/api
```

API tidak menyimpan business logic di frontend. Client seharusnya menggunakan response backend untuk klasifikasi gula darah, rekomendasi, statistik, dan data monitoring.
