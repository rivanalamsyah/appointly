# 🤖 Panduan Pengembang & Instruksi Asisten Claude — Appointly Frontend

Dokumen ini berisi standar pengembangan, tata cara eksekusi, serta konvensi kode untuk asisten AI pada proyek frontend Appointly.

---

## 🚀 Perintah Jalur Pengembang (Development Mode)

Saat menjalankan server pengembangan lokal pada mode latar belakang:

```bash
astro dev --background
```

### Manajemen Proses Server Latar Belakang:
- `astro dev status` — Memeriksa status kesehatan server.
- `astro dev logs`   — Melihat log server secara *real-time*.
- `astro dev stop`   — Menghentikan server pengembangan.

---

## 📐 Standar Kualitas & Aturan Komponen

1. **Prinsip Utama UI**: "Design System First, Page Second". Semua token visual harus memanfaatkan variabel CSS di `src/styles/global.css`.
2. **Aksesibilitas (WCAG AA)**: Setiap tombol ikonik wajib memiliki label `aria-label`, semua modal harus menangani navigasi keyboard `Escape`, dan elemen interaktif memiliki garis fokus terdefinisi (`focus-visible`).
3. **Penyelarasan Tipe Data**: Semua komponen interaktif React Islands harus memanfaatkan antarmuka tipe TypeScript dari `src/types/api.ts`.
4. **Bebas Data Palsu**: Komponen produksi tidak boleh menyertakan statistik atau data tiruan permanen tanpa pemisahan antarmuka API yang jelas.
