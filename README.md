# CS7-SC - Website Kelas + Auth

Website kelas CS7-SC. Frontend Astro statis, API auth Go, PostgreSQL, dan deploy Docker Compose dengan HTTPS otomatis lewat Caddy.

> **Status deploy:** belum live. Baca [`HANDOFF.md`](HANDOFF.md) untuk status terbaru, blocker, dan langkah deploy.

## Stack

- Web: Astro 7, TypeScript strict, Tailwind v4
- Auth: Go 1.27, `net/http`, pgx, Argon2id
- Database: PostgreSQL 17
- Deploy: Docker Compose + Caddy

## Mulai (frontend)

```sh
nvm use
npm ci
npm run dev
```

## Perintah

| Command                      | Action                    |
| :--------------------------- | :------------------------ |
| `npm run dev`                | Dev server (proxy `/api`) |
| `npm run build`              | Build statis ke `dist/`   |
| `npm run lint`               | ESLint                    |
| `npm run format:check`       | Cek format Prettier       |
| `npm run test`               | Unit test Vitest          |
| `npm run check`              | Astro check               |
| `npm run audit`              | Dependency check          |
| `cd server && go test ./...` | Test backend              |
| `cd server && go vet ./...`  | Static check backend      |

## Backend lokal

Butuh Go 1.27+ dan PostgreSQL (atau Docker untuk database). Buat `.env` lokal
berisi `DOMAIN=localhost` dan `POSTGRES_PASSWORD` pilihan Anda, lalu:

```sh
docker compose up -d db
cd server
DATABASE_URL="postgres://cs7:PASSWORD@127.0.0.1:5432/cs7?sslmode=disable" \
APP_ORIGIN="http://localhost:4321" \
COOKIE_SECURE=false \
go run ./cmd/api
```

Buat akun dari terminal lain:

```sh
cd server
DATABASE_URL="postgres://cs7:PASSWORD@127.0.0.1:5432/cs7?sslmode=disable" \
APP_ORIGIN="http://localhost:4321" \
go run ./cmd/api create-user --email pengurus@contoh.id --name "Pengurus" --role admin --generate
```

## Deploy

1. Arahkan A record domain ke IP server.
2. Pastikan port 22, 80, dan 443 terbuka di firewall/security group.
3. Di server, siapkan repo dan `.env`:

   ```sh
   cp .env.example .env
   openssl rand -hex 32   # isi POSTGRES_PASSWORD
   chmod 600 .env
   ```

4. Build frontend, lalu jalankan:

   ```sh
   npm ci
   npm run build
   docker compose up -d --build
   ```

5. Buat akun admin pertama:

   ```sh
   docker compose exec -T api /app/api create-user \
     --email pengurus@contoh.id --name "Pengurus" --role admin --generate
   ```

6. Cek `https://DOMAIN/api/health` dan login di `https://DOMAIN/masuk`.

Reset kata sandi akun:

```sh
docker compose exec -T api /app/api set-password --email pengurus@contoh.id --generate
```

## Aturan

- Baca `AGENTS.md` + `docs/00_SSOT.md` + `docs/01_CONTEXT.md` + `docs/02_RULES.md`.
- UI butuh `DESIGN.md` dulu. Tanpa arah = draft.
- Rahasia hanya di `.env` server. Jangan commit `.env`, password, atau token.
- Cara kontribusi: lihat `CONTRIBUTING.md`.
- Lisensi: closed-source, lihat `LICENSE`.
