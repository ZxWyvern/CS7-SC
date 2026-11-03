# 01 Context - CS7-SC Web + Auth

## Stack Keputusan

- Web: Astro minimal template, output statis (`dist/`).
- Styling: Tailwind v4 (`src/styles/landing.css`, token oklch) untuk landing dan halaman auth.
  `src/styles/global.css` tetap untuk halaman kelas non-landing.
- API: Go 1.27, stdlib `net/http`, `pgx/v5`, Argon2id (`golang.org/x/crypto`).
- Database: PostgreSQL 17, hanya dapat diakses dari jaringan Docker internal.
- Deploy: Docker Compose (`compose.yaml`) + Caddy untuk HTTPS otomatis.
- Struktur:
  - `src/pages/` - route saja, logika minimal
  - `src/layouts/` - shell HTML, SEO, a11y (`LandingLayout` untuk landing)
  - `src/components/landing/` - halaman utama kelas: Header, Hero, Stats
    (Sekilas Kelas), Services (Tentang), Features (Pengurus), Pricing
    (Kegiatan), Cta, Faq, Footer. Konten Bahasa Indonesia.
  - `src/components/ui/` - presentational halaman kelas, props jelas
  - `src/lib/` - pure TS, wajib unit test
  - `src/content/` - (nanti) data kelas sebagai markdown/json, bukan backend
  - `server/cmd/api/` - entrypoint API + CLI admin (`create-user`, `set-password`)
  - `server/internal/` - config, auth (Argon2id/sesi), httpapi, ratelimit, store
  - `server/migrations/` - skema awal PostgreSQL
  - `deploy/` - Caddyfile
  - `public/` - aset statis
- Auth: akun dibuat admin/pengurus via CLI. Sesi opaque 32 byte, hash SHA-256 di database,
  cookie `cs7_session` 7 hari, HttpOnly + Secure + SameSite=Strict.

## Halaman

- `/` Landing kelas CS7-SC
- `/masuk` Login email + kata sandi
- `/akun` Halaman akun (client-side guard, redirect ke `/masuk` bila belum login)
- `/jadwal`, `/galeri` (segera, jangan link mati - R-24)

## API

- `POST /api/auth/login` - `{email, password}` -> user + cookie sesi
- `POST /api/auth/logout` - hapus sesi
- `GET /api/auth/me` - user aktif atau 401
- `GET /api/health` - cek API + database

## Perintah

- `npm run dev` - dev server (proxy `/api` ke `http://127.0.0.1:8080`)
- `npm run build` - build statis
- `npm run check` - astro check
- `npm run lint` / `lint:fix`
- `npm run format:check` / `format`
- `npm run test` - vitest run
- `npm run audit` - npm audit moderate
- `cd server && go test ./...` - unit test backend
- `cd server && go vet ./...` - static check backend
- `docker compose up -d --build` - jalankan db + api + caddy di server
