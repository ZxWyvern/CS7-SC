# 01 Context - CS7-SC Frontend Only

## Stack Keputusan

- Astro minimal template, output statis (`dist/`).
- Styling: Tailwind v4 (`src/styles/landing.css`, token oklch) untuk landing.
  `src/styles/global.css` tetap untuk halaman kelas non-landing.
- Struktur:
  - `src/pages/` - route saja, logika minimal
  - `src/layouts/` - shell HTML, SEO, a11y (`LandingLayout` untuk landing)
  - `src/components/landing/` - port template Homie (Header/Hero/Stats/Services/
    Features/Cta/Pricing/Testimonials/Faq/Footer), konten Inggris ikut template
  - `src/components/ui/` - presentational halaman kelas, props jelas
  - `src/lib/` - pure TS, wajib unit test
  - `src/content/` - (nanti) data kelas sebagai markdown/json, bukan backend
  - `public/` - aset statis (termasuk `public/images/` dari template Homie)
- Backend: tidak ada. Form pakai `mailto:` atau link, bukan POST fiktif.

## Halaman Awal

- `/` Landing Homie (port template, arah di `DESIGN.md`)
- `/jadwal`, `/galeri` (segera, jangan link mati - R-24)

## Perintah

- `npm run dev` - dev server
- `npm run build` - build statis
- `npm run check` - astro check
- `npm run lint` / `lint:fix`
- `npm run format:check` / `format`
- `npm run test` - vitest run
- `npm run audit` - npm audit moderate
