# 02 Rules - Frontend CS7-SC

## Cara Penulisan (Code)

- TypeScript strict, no `any` tanpa alasan di komentar.
- Component Astro: frontmatter tipis, props typed, tanpa fetch backend.
- CSS: token dulu, tanpa magic number. Mobile-first, min 44px tap target.
- Test: setiap fungsi di `src/lib/` wajib `*.test.ts`.
- Commit: kecil, pesan jelas Indonesia/Inggris konsisten.

## Aturan UI/UX (antislop + frontend-design)

- Ikuti `AGENTS.md`: baca `DESIGN.md` dulu (R-37), tanpa arah = draft, bukan deliverable.
- Purpose test: tiap teknik visual tulis alasan 1 baris.
- Larangan keras:
  - em dash (—), badge AI generik, statistik palsu, testimoni fiktif
  - link nav mati, tombol mati, FAQ template
  - kontras di bawah 4.5:1, outline dihapus tanpa ganti, hover-only tanpa tap
  - patch CSS via script, dark mode rusak sebelah
- Liveliness: set dial ENERGY/RHYTHM/MOTION di DESIGN.md, 1 focal point per layar.
- frontend-design: 1 elemen memorable, sisanya disiplin. Rencana token 4-6 hex + tipografi + wireframe ASCII sebelum koding.

## Workflow

- Semua PR wajib hijau: lint, test, deps, build.
- Jangan merge kalau ada FAIL di Delivery Gate antislop.
