# 🛡️ Kebijakan & Arsitektur Keamanan Sistem (Security Model) — Appointly

> Dokumentasi resmi mengenai arsitektur keamanan, isolasi multi-tenant, otorisasi RBAC, perlindungan data sensitif, serta pertahanan jaringan aplikasi **Appointly**.

---

## 📑 Daftar Isi

- [1. Prinsip Keamanan & Ancaman](#1-prinsip-keamanan--ancaman)
- [2. Strategi Isolasi Data Multi-Tenant](#2-strategi-isolasi-data-multi-tenant)
- [3. Pemisahan Domain Pembayaran](#3-pemisahan-domain-pembayaran)
- [4. Autentikasi & Hirarki RBAC](#4-autentikasi--hirarki-rbac)
- [5. Sistem Audit Log Imutabel](#5-sistem-audit-log-imutabel)
- [6. Pertahanan Jaringan & Header HTTP](#6-pertahanan-jaringan--header-http)
- [7. Pembatas Laju API (Rate Limiting)](#7-pembatas-laju-api-rate-limiting)
- [8. Keamanan Webhook & Idempotensi](#8-keamanan-webhook--idempotensi)
- [9. Enkripsi & Pelindungan Data Sensitif](#9-enkripsi--pelindungan-data-sensitif)
- [10. Pelaporan Kerentanan Keamanan](#10-pelaporan-kerentanan-keamanan)

---

## 1. Prinsip Keamanan & Ancaman

Appointly menggunakan arsitektur **Defense-in-Depth (Pertahanan Berlapis)** dan prinsip **Zero-Trust Multi-Tenant Architecture**. Setiap lapisan sistem — mulai dari endpoint HTTP publik hingga kueri basis data paling dasar — memverifikasi konteks tenant, autentikasi sesi, dan integritas jejak audit.

---

## 2. Strategi Isolasi Data Multi-Tenant

- **Konteks Tenant Wajib (`organization_id`)**: Seluruh repositori domain (`Appointment`, `Staff`, `Customer`, `Service`, `Location`, `Resource`, `Payment`, `Subscription`, `AuditLog`) membutuhkan parameter `organization_id` yang terverifikasi.
- **Penolakan Akses Lintas-Tenant (Cross-Tenant Rejection)**: Logika otorisasi mengevaluasi `authCtx.OrgID == target.OrganizationID`. Setiap ancaman atau upaya akses lintas-tenant mengembalikan respons `403 Forbidden` atau `404 Not Found` tanpa membocorkan keberadaan sumber daya milik organisasi lain.
- **Pengujian Terotomatisasi**: Diverifikasi secara otomatis melalui suite tes keamanan pada `internal/usecase/security/cross_tenant_test.go`.

---

## 3. Pemisahan Domain Pembayaran

- **Pembayaran Janji Temu Pelanggan (B2C)**: Diproses melalui abstraksi provider pembayaran terverifikasi (Stripe, Midtrans, Pembayaran di Tempat).
- **Pembayaran Langganan SaaS Organisasi (B2B)**: Dikelola secara terpisah melalui modul billing langganan platform tenant (`Plan`, `Subscription`, `UsageRecords`).
- **Jaminan Keamanan**: Transaksi pembayaran pelanggan publik tidak akan pernah dapat mengubah atau mempengaruhi status langganan SaaS organisasi penyedia jasa.

---

## 4. Autentikasi & Hirarki RBAC

### Pemisahan Identitas & Keanggotaan Organisasi
- **User Account**: Akun pengguna sistem yang menyimpan kredensial login, email, hash kata sandi, dan status sesi.
- **Organization Membership**: Menghubungkan pengguna ke Organisasi dengan Peran tertentu (`Owner`, `Admin`, `Staff`, `Member`).
- **Staff Profile**: Profil operasional staf untuk penjadwalan. Profil staf dapat berdiri sendiri tanpa harus memiliki akun login sistem.

### Matriks Izin Akses RBAC

| Peran (Role) | Kelola Bisnis & Settings | Kelola Staf & Layanan | Lihat Semua Booking | Eksekusi Booking & Reschedule | Lihat Keuangan & Billing | Lihat Audit Log Imutabel |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Owner** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Admin** | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Staff** | ❌ | ❌ | Hanya Ditugaskan / Konfigurasi | ✅ | ❌ | ❌ |
| **Member** | ❌ | ❌ | Milik Sendiri | ✅ | ❌ | ❌ |

---

## 5. Sistem Audit Log Imutabel (`AuditLog`)

### Jaminan Imutabilitas Data Audit
- **Arsitektur Append-Only**: Repositori `AuditRepository` hanya menyediakan metode `Create`, `List`, dan `GetByID`. **Tidak ada fungsi `Update` atau `Delete` dalam kode aplikasi maupun API.**
- **Batasan Basis Data**: Aturan SQL PostgreSQL `BEFORE UPDATE OR DELETE` aktif pada tabel audit log untuk menggagalkan pembaruan data secara langsung.
- **Hak Akses Khusus**: Log audit hanya dapat dibaca oleh pengguna dengan peran `Owner` dan `Admin` yang terverifikasi.

---

## 6. Pertahanan Jaringan & Header HTTP

Setiap respons dari server API Appointly menyertakan header keamanan berlapis:

```http
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; frame-ancestors 'none';
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

---

## 7. Pembatas Laju API (Rate Limiting)

- **Endpoint API Terautentikasi**: Maksimal 120 permintaan / menit per IP.
- **Endpoint Pemesanan Publik**: Maksimal 60 permintaan / menit per IP.
- **Endpoint Autentikasi (`/api/v1/auth/*`)**: Maksimal 20 permintaan / menit per IP (Melindungi dari serangan *Brute-Force*).

---

## 8. Keamanan Webhook & Idempotensi

- Webhook masuk dari penyedia pembayaran atau layanan eksternal wajib memverifikasi tanda tangan HMAC SHA-256 (`X-Signature` / `X-Hub-Signature-256`).
- Kunci idempotensi disimpan dalam Redis/Postgres untuk mencegah serangan pembalasan (*replay attacks*).

---

## 9. Enkripsi & Pelindungan Data Sensitif

- **Enkripsi Kata Sandi**: Kata sandi di-hash menggunakan **Argon2id** / **Bcrypt** dengan *salt* acak kriptografis.
- **Pembersihan Log (Log Sanitization)**: Kata sandi mentah, token JWT, cookie sesi, dan kunci rahasia webhook **dilarang keras dicatat ke dalam log**.

---

## 10. Pelaporan Kerentanan Keamanan

Jika Anda menemukan kerentanan keamanan pada platform Appointly, harap kirimkan laporan lengkap ke tim keamanan kami melalui email: `security@appointly.app`. Tim kami akan merespons dalam waktu 24 jam.
