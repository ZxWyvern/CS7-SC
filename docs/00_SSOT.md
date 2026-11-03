# 00 SSOT - CS7-SC Website Kelas

- Tujuan: website kelas CS7-SC + autentikasi akun pengurus/anggota yang dibuat admin.
- Stack web: Astro 7 (SSG), TypeScript strict, CSS murni + Tailwind v4 untuk landing.
- Stack auth: Go 1.27 (stdlib + pgx), PostgreSQL 17, Docker Compose, Caddy (HTTPS).
- Bahasa: Indonesia.
- Scope saat ini: login, logout, sesi, halaman akun. Tidak ada pendaftaran publik, reset kata sandi via email, atau data siswa.
- Workflow wajib: Lint, Unit Test, Dependency Check, Build via GitHub Actions; backend `go test` + `go vet`.
- SSOT menang: file ini + `01_CONTEXT.md` + `02_RULES.md`. Kalau konflik, SSOT menang.
- No scope creep: jangan tambah fitur di luar auth tanpa persetujuan.
- Security lock: password Argon2id, cookie sesi HttpOnly + Secure + SameSite=Strict, rate limit login, HTTPS wajib. Rahasia hanya di `.env` server.
- Numbers Lock: jangan pajang angka/testimoni tanpa sumber nyata.
