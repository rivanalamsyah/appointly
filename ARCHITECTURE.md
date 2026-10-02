# 🏗️ Dokumen Keputusan Arsitektur Sistem (Architecture Decision Record) — Appointly

> **Versi**: 1.0.0  
> **Status**: Dokumen Resmi Tingkat Produksi  
> **Terakhir Diperbarui**: 2026-10-03  

---

## 📑 Daftar Isi

- [1. Ikhtisar Sistem](#1-ikhtisar-sistem)
- [2. Prinsip Arsitektur Utama](#2-prinsip-arsitektur-utama)
- [3. Stack Teknologi & Rationale](#3-stack-teknologi--rationale)
- [4. Struktur Repositori & Pola Lapisan Kode](#4-struktur-repositori--pola-lapisan-kode)
- [5. Model Domain & Entitas Inti](#5-model-domain--entitas-inti)
- [6. Aturan Isolasi Multi-Tenant](#6-aturan-isolasi-multi-tenant)
- [7. Mesin Ketersediaan Dinamis (Availability Engine)](#7-mesin-ketersediaan-dinamis-availability-engine)
- [8. Mekanisme Pencegahan Double-Booking Konkuren](#8-mekanisme-pencegahan-double-booking-konkuren)
- [9. Desain API & Kontrak Respons](#9-desain-api--kontrak-respons)
- [10. Kebijakan Keamanan & RBAC](#10-kebijakan-keamanan--rbac)
- [11. Arsitektur Pekerja Latar Belakang & Asinkron](#11-arsitektur-pekerja-latar-belakang--asinkron)
- [12. Matriks Keputusan Rekayasa (Decision Log)](#12-matriks-keputusan-rekayasa-decision-log)

---

## 1. Ikhtisar Sistem

**Appointly** adalah platform *Software-as-a-Service* (SaaS) multi-tenant tingkat produksi yang dirancang untuk manajemen janji temu (*appointment*) dan pemesanan online (*online booking*). 

Meskipun vertikal awal difokuskan pada industri jasa seperti Salon, Barbershop, dan Spa, arsitektur sistem dirancang secara **Agnostik Vertikal** (*vertical-agnostic*). Entitas inti menggunakan abstraksi generik seperti `staff`, `service`, `location`, dan `resource` sehingga sistem dapat melayani klinik kesehatan, konsultan hukum, lembaga bimbingan belajar, pusat kebugaran, dan penyedia jasa lainnya tanpa mengubah struktur skema basis data utama.

---

## 2. Prinsip Arsitektur Utama

| Prinsip | Implementasi Teknis |
| :--- | :--- |
| **Isolasi Tenant Mutlak** | Organisasi (`Organization`) berfungsi sebagai tenant utama. Setiap kueri basis data dijamin terisolasi melalui konteks server `organization_id`. |
| **Agnostik Vertikal** | Struktur data generik tanpa kolom khusus industri tertentu di skema basis data inti. |
| **Desain Berbasis Domain (DDD)** | Batasan domain yang jelas (`Auth`, `Appointment`, `Payment`, `Scheduling`, `Subscription`, dll) tanpa *state* bersama yang dapat berubah (*no shared mutable state*). |
| **Arsitektur Berlapis (Layered Architecture)** | Pemisahan tegas: **Handler/Transport** → **Use Case/Application** → **Domain Logic** → **Repository/Infrastructure**. |
| **Transaksional & Idempotent** | Operasi penulisan kritis menggunakan transaksi basis data eksplisit dan kunci idempotensi untuk mencegah efek samping ganda. |
| **Keamanan Terintegrasi (Security-by-Design)** | Autentikasi JWT, otorisasi RBAC, pembatasan laju (*rate-limiting*), dan audit log imutabel aktif secara *default*. |

---

## 3. Stack Teknologi & Rationale

### 3.1 Backend Engine
- **Bahasa Pemrograman**: **Go 1.26+** — Memberikan performa tinggi, efisiensi memori, serta manajemen konkurensi native yang andal melalui *goroutines*.
- **Router HTTP**: **Chi Router** — Ringan, idiomatic Go, dan memiliki sistem composable middleware yang cepat.
- **Basis Data Utama**: **PostgreSQL 16** — Menjamin integritas data berstandar ACID, mendukung tipe data JSONB untuk metadata fleksibel, dan memiliki constraint unik parsial yang kuat.
- **Akses & Kueri SQL**: **sqlc + raw SQL** — Meng-generate kode Go *type-safe* langsung dari berkas query `.sql` tanpa overhead *reflection* ORM.
- **Migrasi Basis Data**: **golang-migrate** — Pengelolaan skema basis data terversi secara otomatis dan eksplisit.
- **Cache & Antrean**: **Redis 7** — Penyimpanan sesi, pembatas laju API, pub/sub, dan manajemen antrean kerja asinkron.
- **Penyimpanan Berkas**: **MinIO (Dev) / AWS S3 (Prod)** — Penyimpanan objek terenkripsi untuk aset gambar dan dokumen.

### 3.2 Frontend Application
- **Framework Utama**: **Astro 5** — Menggunakan *Islands Architecture* untuk menghasilkan output statis HTML cepat dengan hidrasi parsial.
- **Komponen Interaktif**: **React 19** — Digunakan pada komponen kompleks yang membutuhkan interaktivitas tinggi (Kalender interaktif, Wizard Pemesanan Publik, Modal Konfirmasi).
- **Sistem Desain UI**: **Tailwind CSS v4** — Manajemen token visual terpusat (Warna HSL, Tipografi, Spacing, Shadow).
- **State Server & Fetching**: **TanStack Query v5** — Manajemen status server, caching otomatis, dan pembaruan latar belakang.

---

## 4. Struktur Repositori & Pola Lapisan Kode

```text
appointly/
├── backend/
│   ├── cmd/
│   │   ├── api/               # Biner Server API HTTP
│   │   └── worker/            # Biner Pengolah Pekerjaan Latar Belakang
│   ├── internal/
│   │   ├── config/            # Pemuatan Konfigurasi dari Environment
│   │   ├── domain/            # Entitas Domain Murni & Interface Repositori
│   │   ├── usecase/           # Orchestration Logika Bisnis Aplikasi
│   │   ├── handler/           # Transport Layer HTTP (Penerjemah JSON/REST)
│   │   │   └── v1/
│   │   │       └── public/    # API Pemesanan Publik Tanpa Auth
│   │   ├── repository/        # Implementasi Akses Basis Data (Postgres/Redis)
│   │   ├── infrastructure/   # Adaptor Provider Eksternal (Payment, Mail)
│   │   ├── middleware/        # Middleware HTTP (Auth, RBAC, Tenant, Rate Limit)
│   │   └── worker/            # Pengolah Job Antrean Asinkron
│   └── db/
│       ├── migrations/        # Berkas Migrasi SQL (000001 - 000008)
│       └── queries/           # Berkas Kueri SQL sqlc
```

---

## 5. Model Domain & Entitas Inti

### Relasi Entitas Domain (ERD Conceptual)

```text
Organization (Tenant)
  ├── Location (Cabang / Lokasi Fisik)
  ├── Staff (Profil Staf & Jam Kerja)
  ├── Service (Katalog Layanan & Durasi)
  ├── Resource (Ruangan, Alat, Meja)
  ├── Customer (Basis Data Pelanggan)
  ├── Appointment (Transaksi Pemesanan Janji Temu)
  │     └── AppointmentStatusHistory (Riwayat Perubahan Status)
  ├── Payment (Catatan Transaksi Pembayaran)
  └── Subscription (Paket SaaS Organisasi)
```

---

## 6. Aturan Isolasi Multi-Tenant

1. **Atribut `organization_id` Mandatory**: Setiap tabel bisnis wajib memiliki kolom `organization_id` berindeks dengan Foreign Key `NOT NULL`.
2. **Konteks Server Aman**: `organization_id` diekstrak dari token JWT pengguna terautentikasi dan disimpan dalam `context.Context`.
3. **Penolakan Parameter Client**: Nilai `organization_id` yang dikirim dalam *request body* atau *query parameter* oleh klien **DIABAIKAN**.
4. **Verifikasi Repositori**: Setiap klausa `SELECT`, `UPDATE`, dan `DELETE` pada repositori **WAJIB** mencantumkan `WHERE organization_id = $1`.

---

## 7. Mesin Ketersediaan Dinamis (Availability Engine)

Mesin Ketersediaan (*Availability Engine*) berada pada `internal/usecase/availability`. Mesin ini menghitung slot waktu luang secara *real-time* tanpa mengandalkan tabel slot statis yang kaku.

### Tahapan Algoritma Ketersediaan

```text
[Input Kueri] (org_id, service_id, staff_id, location_id, date_range, timezone)
      │
      ▼
┌─────────────────────────────────────────────────────────────┐
1. Verifikasi Layanan (Status Aktif & Durasi + Buffer)       │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
2. Batas Pemesanan Organisasi (Min & Max Notice Hours)        │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
3. Irisan Jam Operasional Lokasi & Shift Staf                 │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
4. Pengurangan Window Waktu (Jam Istirahat, Time-Off, Appt)   │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
5. Verifikasi Ketersediaan Sumber Daya (Resource)             │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
[Output Slot Waktu Siap Pesan (Chronologically Sorted)]
```

---

## 8. Mekanisme Pencegahan Double-Booking Konkuren

Untuk mencegah dua permintaan pemesanan secara bersamaan memilih slot waktu dan staf yang sama (*race condition*):

1. **Penguncian Baris Transaksi (`SELECT FOR UPDATE`)**:
   Saat transaksi booking dimulai, sistem mengunci baris jadwal staf yang relevan dalam blok `BEGIN ... COMMIT`.
2. **Indeks Unik Parsial Basis Data**:
   PostgreSQL menegakkan batasan keunikan pada tabel `appointments`:
   ```sql
   CREATE UNIQUE INDEX idx_appointments_no_double_book 
   ON appointments (organization_id, staff_id, start_time) 
   WHERE status IN ('pending', 'confirmed', 'rescheduled');
   ```

---

## 9. Desain API & Kontrak Respons

Seluruh respons API mengembalikan struktur *envelope* JSON yang konsisten:

### Respons Sukses
```json
{
  "data": {
    "id": "appt_12345",
    "status": "confirmed",
    "start_time": "2026-10-05T09:00:00Z"
  },
  "meta": {
    "request_id": "req-987123",
    "timestamp": "2026-10-03T03:50:00Z"
  }
}
```

### Respons Gagal / Error
```json
{
  "error": {
    "code": "SLOT_UNAVAILABLE",
    "message": "Slot waktu yang dipilih telah dipesan oleh pelanggan lain.",
    "details": []
  }
}
```

---

## 10. Kebijakan Keamanan & RBAC

- **Argon2id Password Hashing**: Kata sandi dienkripsi menggunakan algoritma ramah-memori Argon2id.
- **Hirarki Hak Akses (RBAC)**:
  - `Owner`: Akses penuh organisasi, pengaturan keuangan, dan log audit.
  - `Admin`: Akses manajemen operasional, staf, layanan, dan janji temu.
  - `Staff`: Akses melihat dan mengelola janji temu yang ditugaskan (*assigned*).
- **Headers Keamanan HTTP**: Memasang `X-Content-Type-Options`, `X-Frame-Options`, `Content-Security-Policy`, dan `HSTS`.

---

## 11. Arsitektur Pekerja Latar Belakang & Asinkron

Pekerjaan latar belakang (*background jobs*) dikirim secara asinkron setelah transaksi basis data berhasil commit:

- **Pengiriman Email & Notification**: Notifikasi janji temu dan pengingat (*reminder*).
- **Pemrosesan Webhook**: Mengirim event perubahan status ke URL Webhook tenant dengan strategi pengulangan (*exponential backoff*).

---

## 12. Matriks Keputusan Rekayasa (Decision Log)

| No | Keputusan Teknis | Alasan Utama | Alternatif yang Ditolak |
| :--- | :--- | :--- | :--- |
| 1 | **Chi Router** | Performa tinggi, idiomatik, kompatibel dengan `http.Handler` standar. | Gin / Echo |
| 2 | **sqlc + PostgreSQL** | Kueri aman tipe (*type-safe*), bebas overhead reflection ORM. | GORM / Ent |
| 3 | **Single Schema Multi-Tenancy** | Pemeliharaan migrasi basis data yang mudah dan efisien pada koneksi pool. | Schema Per-Tenant |
| 4 | **Astro + React Islands** | Menghasilkan bundel JavaScript minimal di sisi klien dengan rendering cepat. | Next.js SPA |
