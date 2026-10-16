# 01 Context - CS7-SC Frontend Only

## Stack Keputusan

- Astro minimal template, output statis (`dist/`).
- Styling: `src/styles/global.css` + token CSS di `:root`. Tailwind opsional, belum dipasang.
- Struktur:
  - `src/pages/` - route saja, logika minimal
  - `src/layouts/` - shell HTML, SEO, a11y
  - `src/components/ui/` - presentational, props jelas
  - `src/lib/` - pure TS, wajib unit test
  - `src/content/` - (nanti) data kelas sebagai markdown/json, bukan backend
  - `public/` - aset statis
- Backend: tidak ada. Form pakai `mailto:` atau link, bukan POST fiktif.

## Halaman Awal

- `/` Beranda (ada)
- `/jadwal`, `/galeri` (segera, jangan link mati - R-24)

## Perintah

- `npm run dev` - dev server
- `npm run build` - build statis
- `npm run check` - astro check
- `npm run lint` / `lint:fix`
- `npm run format:check` / `format`
- `npm run test` - vitest run
- `npm run audit` - npm audit moderate
