# Bruno Collection

Koleksi request buat API ini, udah jadi — gak perlu import manual dari OpenAPI spec.

## Cara buka

1. Buka aplikasi Bruno.
2. **Open Collection** → pilih folder ini (`docs/api/bruno`).
3. Kanan atas ada dropdown environment (defaultnya "No Environment") → pilih **Local**.
4. Pastiin server API jalan (`make dev-api` atau F5) dan Postgres nyala (`make db-up`).

## Environment vs Postman

Konsepnya sama kayak Postman environment/variables, cuma beda istilah dikit:

- Postman "Environment" → Bruno "Environment" juga, disimpen sebagai file `.bru` (`environments/Local.bru`), bukan JSON tersembunyi — bisa dibuka/diedit langsung, dan **ke-commit ke git** (beda dari Postman yang biasanya environment disimpen di cloud/lokal aplikasi doang).
- Variable dipanggil sama: `{{baseUrl}}`.
- Environment `Local` udah diisiin `baseUrl = http://localhost:8080`. Kalau port API beda (misal ganti `PORT` di `.env`), edit di sini.

## Cookie (session auth)

API ini pakai **cookie session**, bukan Bearer token — jadi gak ada tab "Authorization" yang perlu diisi manual. Bruno otomatis nyimpen & ngirim cookie kayak browser, sama kayak Postman. Alurnya:

1. Jalanin request **auth/Login** — Bruno otomatis nyimpen cookie `session_token` dari response.
2. Request lain di collection ini (yang butuh login) otomatis kebawa cookie itu, gak perlu setting apa-apa lagi.
3. Jalanin **auth/Logout** buat clear session.

## Urutan coba yang disaranin

1. `system/Health`, `system/Ready` — mastiin server hidup.
2. **`make db-seed`** dulu di terminal (bikin user `admin@example.com` / `changeme123`, otomatis jadi SUPER_ADMIN) — kalau belum pernah dijalanin.
3. `auth/Login` — pake kredensial seed di atas.
4. `auth/Me` — mastiin cookie jalan, balikin data admin.
5. `users/List Users`, `users/Create User`, dst — semua endpoint admin sekarang bisa diakses.
6. `roles/List Roles` — cek role `SUPER_ADMIN` udah ada.
7. `audit-logs/List Audit Logs` — cek semua aksi di atas kecatet.

## Request yang butuh ID manual

Beberapa request (`users/Get User`, `roles/Grant Permission`, dst) punya placeholder `REPLACE_WITH_USER_ID` / `REPLACE_WITH_ROLE_ID` di path atau body — copy ID dari response request sebelumnya (misal `List Users` atau `Create Role`), paste manual. Belum ada scripting auto-chain (bisa ditambah nanti kalau kepake — Bruno support pre-request script buat itu).
