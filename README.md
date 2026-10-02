# 📅 Appointly — Multi-Tenant Appointment & Booking SaaS Platform

> **Platform Manajemen Janji Temu & Pemesanan Online Multi-Tenant Tingkat Produksi (Production-Ready).**  
> *Arsitektur dirancang secara **Agnostik Vertikal** — mendukung Salon, Klinik, Konsultan, Hukum, Pelatih, hingga Lembaga Kursus tanpa perlu mengubah struktur skema basis data utama atau mesin pemesanan.*

---

## 📑 Daftar Isi

- [Fitur Utama](#-fitur-utama)
- [Arsitektur Teknis](#-arsitektur-teknis)
- [Persyaratan Sistem](#-persyaratan-sistem)
- [Panduan Memulai Cepat (Quick Start)](#-panduan-memulai-cepat-quick-start)
- [Perintah Makefile](#-perintah-makefile)
- [Struktur Proyek](#-struktur-proyek)
- [Variabel Lingkungan (Environment Variables)](#-variabel-lingkungan-environment-variables)
- [Modul Domain Aplikasi](#-modul-domain-aplikasi)
- [Keamanan & Isolasi Tenant](#-keamanan--isolasi-tenant)
- [Panduan Kontribusi](#-panduan-kontribusi)
- [Lisensi](#-lisensi)

---

## 🌟 Fitur Utama

- **Isolasi Multi-Tenant Ketat**: Setiap entitas bisnis dipisahkan berdasarkan `organization_id` pada tingkat query basis data dan konteks server.
- **Pencegahan Double-Booking**: Penguncian baris eksplisit (`SELECT FOR UPDATE`) dan Indeks Unik Parsial PostgreSQL menjamin tidak ada konflik pemesanan slot waktu secara bersamaan.
- **Mesin Ketersediaan Dinamis (Dynamic Availability Engine)**: Menghitung slot waktu secara *real-time* berdasarkan jam operasional lokasi, jadwal staf, jeda layanan (*buffer times*), serta ketersediaan sumber daya.
- **Pemisahan Pembayaran**: Membedakan transaksi pembayaran janji temu pelanggan (B2C) dengan pembayaran langganan SaaS organisasi (B2B).
- **Alur Pemesanan Publik Mobile-First**: Antarmuka pemesanan untuk pelanggan publik yang responsif, intuitif, dan bebas dari pengalihan perhatian.
- **Sistem Audit Log Imutabel**: Catatan jejak audit yang tidak dapat diubah (*append-only*) untuk mendokumentasikan setiap aksi penting pengguna dan sistem.
- **Notifikasi & Pekerja Latar Belakang (Background Workers)**: Pengiriman email, SMS/WhatsApp, dan webhook secara asinkron dengan penanganan ulang (*retry*) eksponensial.

---

## 🛠️ Arsitektur Teknis

| Komponen | Teknologi | Keterangan |
| :--- | :--- | :--- |
| **Backend Service** | Go 1.26+, Chi Router | RESTful JSON API berkinerja tinggi |
| **Database Utama** | PostgreSQL 16 | ACID-compliant, JSONB, Indeks Unik Parsial |
| **Penyimpanan Kode SQL** | `sqlc` + raw SQL | Query aman tipe (*type-safe*) tanpa keajaiban ORM |
| **Cache & Queue** | Redis 7 | Sesi, pembatas laju (*rate limiting*), dan antrean kerja |
| **Frontend Platform** | Astro 5 + React 19 Islands | Rendering secepat kilat dengan hidrasi parsial |
| **Sistem Desain UI** | Tailwind CSS v4 | Token visual terpusat & komponen terkonfigurasi |
| **State Server UI** | TanStack Query v5 | Cache otomatis, pembaruan di latar belakang |
| **Penyimpanan Objek** | MinIO / AWS S3 | Penyimpanan berkas media & gambar profil |

---

## 💻 Persyaratan Sistem

Pastikan perangkat lunak berikut telah terinstal pada lingkungan pengembangan Anda:

- **Go**: v1.26 atau lebih baru
- **Node.js**: v20.0.0 atau lebih baru (npm v10+)
- **Docker**: Engine v24+ & Docker Compose v2.20+
- **Make**: (Opsional) untuk menjalankan perintah pintas

---

## 🚀 Panduan Memulai Cepat (Quick Start)

### 1. Salin Konfigurasi Lingkungan
```bash
cp .env.example .env
```

### 2. Jalankan Infrastruktur Layanan (Docker)
```bash
make infra-up
```
*Perintah ini akan menjalankan kontainer PostgreSQL 16, Redis 7, dan MinIO.*

### 3. Jalankan Migrasi Basis Data
```bash
make migrate-up
```

### 4. Jalankan Server API Backend
```bash
make dev-api
```

### 5. Jalankan Frontend Astro (pada terminal terpisah)
```bash
make dev-frontend
```

### 6. Akses Layanan Aplikasi
- **Frontend Web**: [http://localhost:4321](http://localhost:4321)
- **API Backend**: [http://localhost:8080](http://localhost:8080)
- **Spesifikasi API Docs**: [http://localhost:8080/docs](http://localhost:8080/docs)
- **Konsol MinIO**: [http://localhost:9001](http://localhost:9001) *(Kredensial: admin / minioadmin)*

---

## 📋 Perintah Makefile

| Target Command | Deskripsi Perintah |
| :--- | :--- |
| `make help` | Menampilkan seluruh daftar perintah Makefile yang tersedia |
| `make infra-up` | Menjalankan kontainer infrastruktur pendukung (Postgres, Redis, MinIO) |
| `make infra-down` | Menghentikan dan menghapus kontainer infrastruktur |
| `make migrate-up` | Menjalankan seluruh migrasi skema basis data |
| `make migrate-down` | Membatalkan (*rollback*) migrasi skema basis data terakhir |
| `make sqlc-gen` | Menggenerate ulang kode Go type-safe dari berkas SQL |
| `make dev-api` | Menjalankan server API backend dengan *live reload* |
| `make dev-worker` | Menjalankan pekerja latar belakang (*worker*) |
| `make dev-frontend` | Menjalankan server pengembangan Astro frontend |
| `make test-unit` | Menjalankan pengujian unit (*unit tests*) |
| `make test-all` | Menjalankan seluruh pengujian (unit, integrasi, dan konkurensi) |
| `make lint` | Menjalankan analisis statis (*golangci-lint* & *eslint*) |
| `make fmt` | Memformat seluruh berkas kode sumber |
| `make build-all` | Membangun biner produksi API backend dan aset frontend |

---

## 📁 Struktur Proyek

```text
appointly/
├── backend/                  # Layanan API Go & Pengolah Pekerjaan Latar Belakang
│   ├── cmd/                  # Titik masuk biner (api & worker)
│   ├── internal/             # Logika aplikasi internal terisolasi
│   │   ├── config/           # Pemuat konfigurasi lingkungan
│   │   ├── domain/           # Entitas & antarmuka domain murni
│   │   ├── handler/          # Pengendali transport HTTP (REST API v1)
│   │   ├── infrastructure/   # Adaptor sistem eksternal (Payment, Mail, Calendar)
│   │   ├── middleware/       # Middleware HTTP (Auth, Tenant, Security, Rate Limit)
│   │   ├── repository/       # Implementasi repositori basis data (Postgres & Redis)
│   │   ├── usecase/          # Logika bisnis & pengorkestrasian aplikasi
│   │   └── worker/           # Pengolah pekerjaan antrean asinkron
│   └── db/                   # Skema migrasi SQL & kueri sqlc
├── frontend/                 # Aplikasi Frontend Astro + React Islands
│   ├── src/
│   │   ├── components/       # Komponen UI terpakai & fitur modul
│   │   ├── layout/           # Shell aplikasi, Sidebar, Topbar, Drawer
│   │   ├── pages/            # Rute halaman Astro (Dashboard & Booking Publik)
│   │   ├── styles/           # Token desain global & Tailwind CSS
│   │   └── types/            # Definisi tipe TypeScript API
├── infra/                    # Konfigurasi Docker & Reverse Proxy (Caddy/Nginx)
├── docker-compose.yml        # Konfigurasi pengorkestrasian kontainer lokal
├── Makefile                  # Otomatisasi Perintah Pengembang
├── ARCHITECTURE.md           # Dokumentasi Arsitektur Mendalam
├── DEPLOYMENT.md             # Panduan Operasional & Peluncuran Produksi
└── SECURITY.md               # Model & Kebijakan Keamanan Sistem
```

---

## 🔒 Keamanan & Isolasi Tenant

Aplikasi Appointly mengimplementasikan pendekatan **Defense-in-Depth** untuk menjamin isolasi data antar organisasi:
- Konteks `organization_id` diekstrak secara otomatis di tingkat server melalui token JWT terenkripsi.
- Setiap kueri basis data pada repositori memuat batasan eksplisit `WHERE organization_id = $1`.
- Pengujian isolasi lintas-tenant (*cross-tenant tests*) dijalankan secara otomatis pada pipeline CI/CD untuk memastikan tidak ada kebocoran data antar organisasi.

Dokumentasi keamanan lengkap dapat dilihat pada [SECURITY.md](./SECURITY.md).

---

## 🤝 Panduan Kontribusi

1. Buat cabang fitur baru dari `main`:  
   `git checkout -b feat/nama-fitur-anda`
2. Patuhi standar penulisan kode sesuai petunjuk pada [ARCHITECTURE.md](./ARCHITECTURE.md).
3. Tulis pengujian unit (*unit test*) untuk setiap logika bisnis baru.
4. Jalankan perintah verifikasi sebelum melakukan commit:  
   `make lint fmt test-unit`
5. Buat *Pull Request* (PR) dengan deskripsi perubahan yang jelas dan terstruktur.

---

## 📄 Lisensi

**Proprietary & Confidential** — Hak Cipta Dilindungi Undang-Undang.
