# AGENTS.md - CS7-SC (Frontend Only, Astro)

Baca `docs/00_SSOT.md` + `docs/01_CONTEXT.md` + `docs/02_RULES.md` dulu. SSOT menang.

## Router Skill

- UI/visual: pakai `frontend-design` untuk arah + `antislop` + `antislop-ui` sebagai filter.
  Global sudah ada di `~/.config/opencode/skills/`:
  `antislop`, `antislop-ui`, `antislop-copywriting`, `antislop-human`,
  `antislop-layoutmobile`, `frontend-design`.
- Copy/text: `antislop-copywriting`
- Aksesibilitas: `antislop-human` (kontras + keyboard + focus)
- Mobile/responsive: `antislop-layoutmobile`
- Cari docs library: `find-docs` (`npx ctx7@latest ...`)
- Arsitektur backend: JANGAN. Frontend only sampai disetujui.

## Aturan Main

- Bahasa: Astro + TS strict. Tanpa backend.
- Sebelum UI: isi `DESIGN.md` atau output = draft without direction (ENERGY 1/RHYTHM 1/MOTION 1).
- Sebelum aset (logo/avatar/angka/testimoni/nav): tanya dulu atau placeholder jujur `[LOGO]`, `[REAL DATA]`.
- Setiap PR: `npm run lint`, `npm run format:check`, `npm run test`, `npm run build` harus hijau.
- Struktur: pages tipis, layouts shell, components/ui presentational, lib pure + test.

<!-- antislop:start -->

## antislop

Mode: DURING. Untuk UI, baca DESIGN.md dulu lalu filter antislop + frontend-design.
Skills global:

- UI/visual: `frontend-design` + `antislop-ui`
- Copy: `antislop-copywriting`
- People: `antislop-human`
- Mobile: `antislop-layoutmobile`
  Gate wajib sebelum deliver: Delivery Gate antislop PASS semua dengan bukti.

<!-- antislop:end -->
