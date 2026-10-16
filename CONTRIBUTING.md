# Berkontribusi ke CS7-SC

Terima kasih mau bantu website kelas. Baca dulu `docs/00_SSOT.md`,
`docs/01_CONTEXT.md`, `docs/02_RULES.md`, dan `AGENTS.md`. SSOT menang.

Proyek ini **closed-source** (lihat `LICENSE`). Jangan sebar kode keluar
tanpa izin tertulis.

## Prasyarat

- Node 22 (`nvm use`), npm terbaru
- `npm ci` setelah pull

## Alur Kerja Frontend-Only

1. Ambil issue / bikin branch dari `master`:
   `git checkout -b feat/nama-singkat` atau `fix/nama-singkat`
2. Commit setiap perubahan di local dulu sebelum push ke origin.
   Satu commit = satu tujuan, pesan jelas. Push ke `origin` hanya
   setelah semua cek lokal hijau.
3. Frontend saja, tanpa backend. Jangan tambah auth, database, API server.
4. Kalau ubah UI, isi `DESIGN.md` dulu. Tanpa arah = draft, bukan deliverable.
5. Aset (logo, foto, angka, testimoni, nav): tanya dulu atau pakai
   placeholder jujur `[LOGO]` / `[REAL DATA]`. Jangan karang.
6. Struktur:
   - `src/pages/` route tipis
   - `src/layouts/` shell + SEO + a11y
   - `src/components/ui/` presentational
   - `src/lib/` pure TS + wajib `*.test.ts`

## Cek Lokal Wajib Hijau

```sh
npm run lint
npm run format:check
npm run test
npm run build
```

Perbaiki dengan `npm run lint:fix` dan `npm run format` bila perlu.

## Aturan UI/UX Singkat

- `frontend-design` untuk arah, `antislop` sebagai filter.
- Tulis alasan 1 baris tiap keputusan visual.
- Dilarang: em dash, statistik palsu, testimoni fiktif, link/tombol mati,
  kontras di bawah 4.5:1, hover-only tanpa tap, patch CSS via script.
- Mobile 360px wajib rapi, tap target min 44px, tanpa overflow horizontal.

## Pull Request

- PR kecil, satu tujuan, deskripsi jelas + screenshot bila ubah UI.
- Checklist PR:
  - [ ] `lint`, `format:check`, `test`, `build` hijau lokal
  - [ ] CI hijau (lint, test, deps, build)
  - [ ] Delivery Gate antislop PASS dengan bukti bila sentuh UI
  - [ ] Tidak ada data palsu atau klaim tanpa sumber
- Maintainer boleh minta revisi. Jangan merge sendiri bila bukan maintainer.
- Jangan commit `dist/`, `node_modules/`, `.env`.

## Melapor Masalah

Sertakan langkah reproduksi, URL/halaman, screenshot, versi browser + OS.
