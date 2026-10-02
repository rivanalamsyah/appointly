# 🎨 Appointly — Dokumentasi Frontend Platform

> **Aplikasi Frontend Berbasis Astro 5 + TypeScript + React 19 Islands + Tailwind CSS v4.**  
> Dirancang dengan standar antarmuka B2B SaaS modern, responsif *mobile-first*, mematuhi standar aksesibilitas WCAG AA, dan siap pakai untuk produksi.

---

## 📑 Daftar Isi

- [Arsitektur & Komponen](#-arsitektur--komponen)
- [Struktur Direktori Frontend](#-struktur-direktori-frontend)
- [Perintah Pengembangan (CLI Commands)](#-perintah-pengembangan-cli-commands)
- [Sistem Desain & Token Visual](#-sistem-desain--token-visual)
- [Integrasi API & State Server](#-integrasi-api--state-server)

---

## 🏗️ Arsitektur & Komponen

Frontend Appointly mengadopsi **Astro Islands Architecture**:
- **Output Statis & SSR**: Halaman utama, pemasaran, dan shell dashboard di-render sebagai HTML super cepat.
- **Hidrasi Parsial React**: React 19 Islands hanya digunakan pada komponen interaktif yang kompleks seperti Mesin Pemesanan Publik (*Booking Wizard*), Kalender Interaktif, Pengelola Layanan, dan Modal Konfirmasi.

---

## 📂 Struktur Direktori Frontend

```text
frontend/
├── src/
│   ├── components/
│   │   ├── layout/       # Shell aplikasi, Sidebar responsif, Topbar, Drawer
│   │   ├── ui/           # Komponen primitif terpakai (Button, Modal, Table, Switch, Tooltip)
│   │   ├── features/     # Komponen fitur modul domain (Appointments, Calendar, Services, Staff)
│   │   └── marketing/    # Komponen halaman utama pemasaran
│   ├── lib/              # Api Client (ky), TanStack Query Client, Utilities
│   ├── pages/            # Rute Halaman Astro
│   │   ├── dashboard/    # Dashboard Admin Organisasi (Terautentikasi)
│   │   ├── book/[slug]/  # Portal Pemesanan Pelanggan Publik
│   │   └── index.astro   # Landing Page Pemasaran
│   ├── styles/           # Token Desain Global & Custom CSS Variables
│   └── types/            # Definisi Tipe TypeScript API & Domain
├── astro.config.mjs      # Konfigurasi Astro (Adapter Node & Integrasi React)
├── tailwind.config.mjs   # Konfigurasi Tailwind CSS v4
├── tsconfig.json         # Konfigurasi TypeScript
└── package.json          # Dependensi & Naskah Perintah
```

---

## 💻 Perintah Pengembangan (CLI Commands)

Seluruh perintah dijalankan dari folder `frontend/`:

| Perintah | Deskripsi Tindakan |
| :--- | :--- |
| `npm install` | Menginstal seluruh dependensi paket frontend |
| `npm run dev` | Menjalankan server pengembangan lokal di `http://localhost:4321` |
| `npm run build` | Kompilasi build produksi ke folder `./dist/` |
| `npm run preview` | Pratinjau hasil build produksi secara lokal |
| `npm run typecheck` | Menjalankan verifikasi tipe TypeScript (`tsc --noEmit`) |
| `npm run lint` | Menjalankan analisis kualitas kode dengan ESLint |

---

## 🎨 Sistem Desain & Token Visual

Token desain dikelola secara terpusat pada [`src/styles/global.css`](file:///d:/appointly/frontend/src/styles/global.css) menggunakan variabel CSS:
- **Tipografi**: Display Font (`Outfit`), Body Font (`Inter`), Monospace (`JetBrains Mono`).
- **Palet Warna**: Mode gelap default (`#0f0f13` base, `#16161d` raised cards) dengan gradien Indigo/Violet.
- **Aksesibilitas**: Kontras warna memenuhi standar WCAG AA dan dilengkapi indikator fokus keyboard yang jelas.

---

## 🔗 Integrasi API & State Server

- **API Client**: Menggunakan `ky` fetch wrapper terkonfigurasi di `src/lib/api-client.ts` yang menginjeksi token autentikasi JWT dan header tenant secara otomatis.
- **State Server**: Menggunakan `TanStack Query v5` di `src/lib/query-client.ts` untuk pembaruan data otomatis, caching, dan penanganan status *loading/error/empty*.
