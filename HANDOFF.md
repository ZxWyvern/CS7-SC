# Handoff - Auth & Deploy CS7-SC

Terakhir diperbarui: 2026-11-03. Baca bersama `docs/00_SSOT.md`, `docs/01_CONTEXT.md`, dan `docs/02_RULES.md`. SSOT tetap menang.

## Status

- Kode auth, frontend login, Docker, dan Caddy sudah selesai secara lokal.
- **Belum di-deploy.** Tidak ada layanan yang berjalan di server.
- Server target: `pandu@13.221.130.9`, port `22`.
- Port `22` masih timeout dari IP publik `125.162.255.43` meski whitelist diklaim aktif.
- Port `443` di `13.221.130.9` mengembalikan sertifikat Amazon Chime (`CN=ice.m1.ue1.app.chime.aws`). Verifikasi dulu apakah IP tersebut benar milik server.
- Domain belum diberikan, jadi HTTPS/Caddy belum bisa diuji.
- Akun admin pertama belum dibuat.
- Password SSH pernah dikirim di chat. Jangan simpan di repo; rotasi setelah akses berhasil dan pasang SSH key.

## Yang Sudah Selesai

- Go API: login, logout, session `HttpOnly` + `Secure` + `SameSite=Strict`, rate limit, Argon2id, CLI admin.
- Frontend: `/masuk`, `/akun`, validasi client, redirect aman, proxy dev `/api`.
- Deploy: `compose.yaml` (Postgres 17 + API + Caddy), `deploy/Caddyfile`, `.env.example`.
- Migrasi database: `server/migrations/001_init.sql`.
- Dokumentasi: `README.md`, `AGENTS.md`, `docs/00_SSOT.md`, `docs/01_CONTEXT.md`, `docs/02_RULES.md`.
- Verifikasi lokal:
  - `go test ./...` dan `go vet ./...` lulus.
  - `npm run lint`, `npm run test` (9 lulus), `npm run build` (3 halaman), `npm run format:check` lulus.
  - Caddyfile divalidasi dengan Caddy 2.11.7: `Valid configuration`.
  - Alur login/logout/redirect diuji headless Chrome memakai fake API.

Catatan: perubahan working tree di `DESIGN.md`, `src/components/landing/*`, dan `src/styles/landing.css` sudah ada sebelum pekerjaan auth. Jangan revert tanpa konfirmasi.

## Jalankan Lokal

Frontend:

```sh
npm ci
npm run dev
```

Backend butuh PostgreSQL. Dari root repo:

```sh
cp .env.example .env
# isi DOMAIN=localhost dan POSTGRES_PASSWORD
docker compose up -d db
```

Lalu:

```sh
cd server
DATABASE_URL="postgres://cs7:PASSWORD@127.0.0.1:5432/cs7?sslmode=disable" \
APP_ORIGIN="http://localhost:4321" \
COOKIE_SECURE=false \
go run ./cmd/api
```

Buat akun:

```sh
cd server
DATABASE_URL="postgres://cs7:PASSWORD@127.0.0.1:5432/cs7?sslmode=disable" \
APP_ORIGIN="http://localhost:4321" \
go run ./cmd/api create-user --email EMAIL --name "NAMA" --role admin --generate
```

## Langkah Deploy

1. Pastikan IP publik server. Di server jalankan `curl -4 ifconfig.me`; cocokkan dengan `13.221.130.9`.
2. Buka TCP `22`, `80`, `443` untuk `125.162.255.43` di firewall OS dan security group AWS.
3. Arahkan A record domain ke IP server, lalu tunggu propagasi DNS.
4. Install Docker + Compose plugin jika belum ada. Sesuaikan dengan distro; contoh Ubuntu:

   ```sh
   sudo apt-get update
   sudo apt-get install -y docker.io docker-compose-v2
   sudo systemctl enable --now docker
   sudo usermod -aG docker pandu
   ```

   Logout/login ulang agar grup `docker` aktif.

5. Upload repo tanpa `node_modules`, `.git`, dan file rahasia. Sertakan `dist/` hasil build lokal jika server tidak punya Node.js.
6. Di direktori deploy server:

   ```sh
   cp .env.example .env
   openssl rand -hex 32
   # isi DOMAIN dan POSTGRES_PASSWORD di .env
   chmod 600 .env
   docker compose up -d --build
   ```

7. Cek status dan log:

   ```sh
   docker compose ps
   docker compose logs -f api caddy
   curl -fsS https://DOMAIN/api/health
   ```

8. Buat admin pertama, lalu simpan password sekali tampil:

   ```sh
   docker compose exec -T api /app/api create-user \
     --email EMAIL --name "NAMA" --role admin --generate
   ```

9. Verifikasi end-to-end:
   - Buka `https://DOMAIN/masuk`, login, pastikan redirect ke `/akun`.
   - Cek cookie `cs7_session` punya `HttpOnly`, `Secure`, `SameSite=Strict`.
   - Logout, lalu buka `/akun` lagi dan pastikan diarahkan ke `/masuk`.
   - Cek `https://DOMAIN/api/health` menghasilkan `{"status":"ok"}`.

## Endpoint & CLI

- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/auth/me`
- `GET /api/health`
- `docker compose exec -T api /app/api create-user --email EMAIL --name "NAMA" --role admin --generate`
- `docker compose exec -T api /app/api set-password --email EMAIL --generate`

## Risiko / Hal yang Belum Teruji

- API Go belum diuji terhadap PostgreSQL asli; baru diuji dengan fake store di unit test.
- `/akun` adalah halaman statis dengan guard client-side. Data sensitif tetap hanya dari API.
- Rate limit disimpan di memori; cocok untuk satu instance API, reset saat restart.
- Port database dipublish ke `127.0.0.1:5432` untuk kebutuhan development. Hapus mapping tersebut di server produksi jika tidak diperlukan.
- Caddy butuh port `80` dan `443` benar-benar bebas. Port `443` saat ini menunjukkan layanan Amazon Chime, jadi wajib dicek sebelum deploy.
- `dist/` tidak masuk git. Build ulang sebelum deploy atau upload artefak hasil build.

## File Kunci

- `compose.yaml` - stack produksi.
- `deploy/Caddyfile` - HTTPS, reverse proxy `/api`, static `/srv`.
- `.env.example` - variabel wajib.
- `server/cmd/api/main.go` - startup API.
- `server/cmd/api/cli.go` - CLI akun.
- `server/internal/auth/` - password, session, validasi.
- `server/internal/httpapi/` - route, handler, keamanan.
- `server/internal/store/store.go` - akses PostgreSQL.
- `server/migrations/001_init.sql` - skema.
- `src/components/ui/LoginForm.astro` - form login.
- `src/lib/login.ts` - validasi dan redirect aman.
- `src/pages/masuk.astro` - halaman login.
- `src/pages/akun.astro` - halaman akun.
