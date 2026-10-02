# 🚀 Panduan Peluncuran Produksi & Operasional (Deployment Runbook) — Appointly

> **Dokumen Operasional Resmi** untuk melakukan *deployment*, manajemen kontainer, pengelolahan basis data, dan prosedur pemulihan bencana (*disaster recovery*) pada lingkungan Staging dan Produksi.

---

## 📑 Daftar Isi

- [1. Topologi Arsitektur Lingkungan Produksi](#1-topologi-arsitektur-lingkungan-produksi)
- [2. Manajemen Variabel Lingkungan & Rahasia](#2-manajemen-variabel-lingkungan--rahasia)
- [3. Pengorkestrasian Kontainer Docker](#3-pengorkestrasian-kontainer-docker)
- [4. Strategi Migrasi Basis Data](#4-strategi-migrasi-basis-data)
- [5. Pemeriksaan Kesehatan & Graceful Shutdown](#5-pemeriksaan-kesehatan--graceful-shutdown)
- [6. Manajemen Pekerja Latar Belakang & Idempotensi](#6-manajemen-pekerja-latar-belakang--idempotensi)
- [7. Standar Observabilitas & Logging](#7-standar-observabilitas--logging)
- [8. Strategi Cadangan Basis Data & Pemulihan Bencana](#8-strategi-cadangan-basis-data--pemulihan-bencana)
- [9. Prosedur Rollback Tanpa Downtime](#9-prosedur-rollback-tanpa-downtime)
- [10. Checklist Kesiapan Produksi (Production Readiness)](#10-checklist-kesiapan-produksi-production-readiness)

---

## 1. Topologi Arsitektur Lingkungan Produksi

```text
                         ┌────────────────────────────────┐
                         │   TLS / HTTPS (Port 80/443)    │
                         │   Caddy / Nginx Reverse Proxy   │
                         └───────────────┬────────────────┘
                                         │
                 ┌───────────────────────┴───────────────────────┐
                 │                                               │
                 ▼                                               ▼
   ┌───────────────────────────┐                   ┌───────────────────────────┐
   │    Astro Frontend App     │                   │       Go API Server       │
   │  Node Standalone Runner   │                   │     Chi / Slog Runtime    │
   │        (Port 4321)        │                   │        (Port 8080)        │
   └───────────────────────────┘                   └─────────────┬─────────────┘
                                                                 │
                                     ┌───────────────────────────┼───────────────────────────┐
                                     │                           │                           │
                                     ▼                           ▼                           ▼
                       ┌──────────────────────────┐  ┌───────────────────────┐  ┌──────────────────────────┐
                       │  PostgreSQL 16 (Primary) │  │  Redis 7 (Cache/RL)   │  │  Object Storage (S3/GCS) │
                       │    Multi-Tenant Data     │  │   Sessions & Limits   │  │   Customer/Staff Media   │
                       └──────────────────────────┘  └───────────────────────┘  └──────────────────────────┘
```

---

## 2. Manajemen Variabel Lingkungan & Rahasia

### ⚠️ Aturan Penting Keamanan Produksi
1. **Dilarang keras menyimpan berkas `.env` atau teks kunci rahasia (*secret*) di repositori Git.**
2. Gunakan penyimpan rahasia terenkripsi seperti AWS Secrets Manager, GCP Secret Manager, Vault, atau GitHub Repository Secrets.
3. Kunci rahasia minimal yang harus digenerate secara acak:
   - `JWT_SECRET`: Minimal 32-byte string acak (`openssl rand -base64 32`).
   - `DB_PASSWORD`: Kata sandi basis data yang kuat (minimal 20 karakter).
   - `REDIS_PASSWORD`: Kata sandi Redis yang kuat (minimal 20 karakter).

### Matriks Parameter Environment

| Variabel | Nilai Produksi Direkomendasikan | Deskripsi |
| :--- | :--- | :--- |
| `APP_ENV` | `production` | Mengaktifkan mode keamanan produksi & mematikan endpoint debug. |
| `PORT` | `8080` | Port HTTP internal untuk server API Go. |
| `DB_HOST` | `postgres` / Endpoint RDS | Hostname basis data PostgreSQL. |
| `DB_PORT` | `5432` | Port basis data PostgreSQL. |
| `DB_NAME` | `appointly_prod` | Nama basis data produksi. |
| `DB_USER` | `appointly` | User basis data master dengan hak akses terisolasi. |
| `DB_PASSWORD` | `<SECRET>` | Kata sandi akses basis data produksi. |
| `DB_SSL_MODE` | `require` / `verify-full` | Memaksa enkripsi SSL/TLS pada koneksi basis data. |
| `REDIS_HOST` | `redis` / Endpoint ElastiCache | Hostname Redis server. |
| `REDIS_PORT` | `6379` | Port Redis server. |
| `REDIS_PASSWORD` | `<SECRET>` | Kata sandi autentikasi Redis. |
| `JWT_SECRET` | `<SECRET>` | String minimal 32 karakter untuk enkripsi token JWT. |
| `CORS_ALLOWED_ORIGINS` | `https://appointly.app` | Daftar domain terverifikasi yang diizinkan memanggil API. |
| `LOG_FORMAT` | `json` | Format log terstruktur JSON untuk Datadog / CloudWatch. |
| `LOG_LEVEL` | `info` | Tingkat Detail Log (`info`, `warn`, `error`). |

---

## 3. Pengorkestrasian Kontainer Docker

### Build Multi-Stage Docker
- **Backend Image**: `golang:1.26-alpine` (builder) → `alpine:3.21` (runner) dengan pengguna non-root `appointly:10001` (Ukuran biner ~18MB).
- **Frontend Image**: `node:22-alpine` (builder) → `node:22-alpine` (runner) dengan pengguna non-root `appointly:10001`.

### Perintah Peluncuran Staging & Produksi

#### 1. Menjalankan Lingkungan Staging
```bash
# Salin dan sesuaikan kredensial lingkungan staging
cp .env.example .env

# Jalankan kontainer produksi
docker compose -f docker-compose.prod.yml up -d --build

# Verifikasi status kesehatan kontainer
docker compose -f docker-compose.prod.yml ps
curl -i http://localhost/healthz
```

#### 2. Peluncuran Pembaruan Tanpa Downtime (*Zero-Downtime Rolling Deployment*)
```bash
# Tarik kode terbaru dari cabang main
git pull origin main

# Build image Docker baru tanpa cache
docker compose -f docker-compose.prod.yml build --no-cache

# Jalankan migrasi basis data terlebih dahulu
docker compose -f docker-compose.prod.yml run --rm migrations

# Perbarui layanan secara bergantian (rolling restart)
docker compose -f docker-compose.prod.yml up -d --no-deps --scale backend=2 --no-recreate
docker compose -f docker-compose.prod.yml up -d --no-deps backend frontend proxy
```

---

## 4. Strategi Migrasi Basis Data

Seluruh berkas migrasi tersimpan pada folder `backend/db/migrations/` dengan urutan penomoran terstruktur (`000001` hingga `000008`).

### Panduan Eksekusi Migrasi
1. **Kompatibilitas Mundur**: Setiap migrasi skema **WAJIB** kompatibel dengan kode backend yang sedang berjalan di produksi.
2. **Penambahan Kolom**: Tambahkan kolom baru sebagai `NULL` atau sertakan nilai `DEFAULT` agar tidak merusak kueri yang sedang aktif.

```bash
# Jalankan seluruh migrasi yang belum dieksekusi
docker run --rm -v $(pwd)/backend/db/migrations:/migrations migrate/migrate:v4.17.0 \
  -path=/migrations \
  -database="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" \
  up
```

---

## 5. Pemeriksaan Kesehatan & Graceful Shutdown

### Endpoint Probes
- `GET /health` / `GET /livez`: Mengembalikan status HTTP `200 OK` dengan JSON `{"status":"ok"}` jika proses server aktif (Liveness probe).
- `GET /ready` / `GET /readyz`: Mengembalikan status HTTP `200 OK` dengan JSON `{"status":"ready"}` jika koneksi Postgres & Redis berfungsi normal (Readiness probe).

### Prosedur Graceful Shutdown
Server API Go mendengarkan sinyal penghentian sistem `SIGINT` dan `SIGTERM`:
1. Server berhenti menerima koneksi HTTP masuk baru secara seketika.
2. Permintaan HTTP yang sedang berjalan diberikan tenggat waktu (*grace period*) selama **30 detik** untuk menyelesaikan eksekusi.
3. Koneksi basis data dan Redis dikosongkan dan ditutup secara aman setelah permintaan aktif selesai.

---

## 6. Manajemen Pekerja Latar Belakang & Idempotensi

1. **Jaminan Penyampaian Setidaknya Sekali (*At-Least-Once Delivery*)**: Pekerja latar belakang pengiriman email, SMS, dan webhook beroperasi dengan jaminan penyampaian *at-least-once*.
2. **Perlindungan Idempotensi**:
   - Pemrosesan Webhook memeriksa tabel `webhook_deliveries` untuk mencegah pengolahan ulang event ganda.
   - Webhook pembayaran menggunakan transaksi `UPSERT` idempotent pada tabel `payments` dan `appointments`.

---

## 7. Standar Observabilitas & Logging

Seluruh log diproduksi dalam format terstruktur JSON menggunakan `log/slog` bawaan Go:

```json
{
  "time": "2026-10-03T03:50:00Z",
  "level": "INFO",
  "msg": "appointment created successfully",
  "request_id": "req-987123",
  "org_id": "org-luxe-salon",
  "appointment_id": "appt-45678"
}
```

---

## 8. Strategi Cadangan Basis Data & Pemulihan Bencana

### Jadwal Cadangan Otomatis
- **Pencadangan Penuh Harian**: Dieksekusi setiap malam pukul 02:00 UTC menggunakan `pg_dump`.
- **Transaction Logs (WAL)**: Diarsip ke AWS S3 / GCS dengan kebijakan retensi 30 hari untuk *Point-In-Time Recovery* (PITR).

```bash
# Perintah Pencadangan Manual
docker exec -t appointly-prod-postgres pg_dump -U appointly -d appointly_prod -Fc > backup_$(date +%Y%m%d_%H%M%S).dump

# Prosedur Pemulihan (Restore)
docker exec -i appointly-prod-postgres pg_restore -U appointly -d appointly_prod < backup_20261003_020000.dump
```

---

## 9. Prosedur Rollback Tanpa Downtime

Jika terjadi kegagalan pada peluncuran versi baru:

```bash
# 1. Kembalikan versi kontainer ke tag rilis sebelumnya
docker compose -f docker-compose.prod.yml down
git checkout tags/v1.0.0
docker compose -f docker-compose.prod.yml up -d

# 2. Batalkan migrasi basis data (jika diperlukan)
docker run --rm -v $(pwd)/backend/db/migrations:/migrations migrate/migrate:v4.17.0 \
  -path=/migrations \
  -database="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" \
  down 1
```

---

## 10. Checklist Kesiapan Produksi (Production Readiness)

- [x] **Security Headers**: Header `CSP`, `HSTS`, `X-Content-Type-Options` aktif pada API server.
- [x] **Pembatas Laju (Rate Limiting)**: Terpasang pada rute Auth (20 req/min), Publik (60 req/min), dan API Terproteksi (120 req/min).
- [x] **Indeks Basis Data**: Indeks unik `idx_appointments_no_double_book` terverifikasi aktif.
- [x] **Pengguna Kontainer Non-Root**: Pengguna Docker dibatasi pada ID `10001`.
- [x] **CI/CD Automation**: Pipeline GitHub Actions terkonfigurasi untuk pemeriksaan sintaks, tipe, linting, unit test, dan pembuatan Docker image.
