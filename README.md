# LOST & FOUND SCHOOL

**“Find it. Report it. Return it.”**

Platform digital sekolah untuk melaporkan, mencari, dan mengembalikan barang yang hilang.
Stack: **Golang + Gin + GORM + MySQL/SQLite + HTML/CSS/JavaScript** dengan tema **dark minimalist**.

## Problem

Di lingkungan sekolah, barang hilang biasanya diumumkan via grup WhatsApp atau manual sehingga:
- Informasi mudah tenggelam, sulit dicari per kategori
- Tidak ada database terpusat & status barang
- Tidak ada verifikasi, riwayat, dan notifikasi

## Solution

Satu platform terpusat dengan alur **Report → Verify → Match → Claim → Return**:
siswa/guru/staf melaporkan & mencari barang, admin verifikasi, sistem memberi skor
kecocokan (matching), klaim direview, barang dikembalikan.

## Features

- Public: landing page, statistik live, laporan terbaru, search + filter + sort + pagination, detail publik
- User: register/login (JWT cookie + bcrypt), dashboard, buat/edit/hapus laporan (milik sendiri),
  upload foto (JPG/PNG/WEBP ≤5MB, nama unik), klaim barang ditemukan, notifikasi + badge, profil + ganti password,
  rekomendasi "Kemungkinan Barang Cocok" (disclaimer included)
- Admin: dashboard statistik + grafik per bulan, kelola laporan (verify/reject wajib alasan),
  kelola klaim (approve/reject), matching (run/review/confirm), kelola user (role/status),
  CRUD kategori (proteksi kategori terpakai), activity log, pengaturan
- Matching service (`services/matching_service.go`): Nama 25% + Kategori 20% + Warna 15% +
  Merek 15% + Ciri 15% + Lokasi 5% + Tanggal 5%; ≥70 cocok, 40–69 mungkin, <40 rendah

## Tech Stack

| Layer | Teknologi |
|---|---|
| Backend | Go 1.23+, Gin, GORM |
| DB | MySQL (produksi) / SQLite pure-Go fallback (dev otomatis) |
| Auth | JWT (`Authorization: Bearer` / cookie `token`), bcrypt |
| Frontend | HTML5 + CSS custom dark + Vanilla JS (fetch API) |

## Installation

```bash
go mod tidy
go build -o lost-found-school .
```

## Database Setup

MySQL:
```sql
CREATE DATABASE lost_found_school;
```
Atur `.env` (`DB_DIALECT=mysql`, dst.). Tanpa MySQL, aplikasi otomatis memakai
`data.db` (SQLite) — migration via GORM AutoMigrate + seed otomatis.

## Environment

Salin `.env.example` ke `.env`:

```env
APP_ENV=development
APP_PORT=8080
DB_DIALECT=auto
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=lost_found_school
JWT_SECRET=change-this-secret
```

## Run Project

```bash
go run main.go
# atau
./lost-found-school
```

Server: `http://localhost:8080`

## Demo Account

| Role | Email | Password |
|---|---|---|
| Admin | `admin@school.local` | `Admin123!` |
| User | `user@school.local` | `User123!` |

## API Documentation

Format sukses: `{success:true,message,data}` + `meta{page,limit,total,pages}` untuk list.
Error: `{success:false,message}`.

Auth: `POST /api/register`, `POST /api/login`, `POST /api/logout` (auth),
`GET/PUT /api/profile`, `PUT /api/profile/password`.

Reports: `GET /api/reports?search=&type=LOST|FOUND&category_id=&status=&location=&sort=newest|oldest|az|za&page=&limit=`,
`GET /api/reports/:id`, `POST /api/reports` (auth, multipart),
`PUT/DELETE /api/reports/:id` (owner/admin), `GET /api/reports/lost|found`, `GET /api/my-reports` (auth).

Categories: `GET /api/categories`, `POST/PUT/DELETE /api/categories[/:id]` (admin).

Claims: `POST /api/claims` (auth), `GET /api/my-claims`, `GET /api/claims/:id`,
`GET /api/admin/claims`, `PUT /api/admin/claims/:id` (admin).

Matching: `GET /api/matches[/:id]`, `POST /api/matches/run` (admin),
`PUT /api/matches/:id` (admin), `GET /api/my-matches` (auth).

Notifications: `GET /api/notifications`, `GET /api/notifications/unread-count`,
`PUT /api/notifications/:id/read`, `PUT /api/notifications/read-all`.

Admin: `GET /api/admin/dashboard`, `GET /api/admin/reports`,
`PUT /api/admin/reports/:id/status`, `DELETE /api/admin/reports/:id`,
`GET/PUT/DELETE /api/admin/users[/:id]`, `GET /api/admin/activity-logs`,
`GET/PUT /api/admin/settings`. Publik: `GET /api/stats`.

Status laporan: `PENDING VERIFIED PROCESSING MATCHED RETURNED REJECTED CANCELLED`.
Status klaim: `PENDING REVIEWING APPROVED REJECTED CANCELLED`.
Status match: `PENDING REVIEWING CONFIRMED REJECTED`.

## Folder Structure

```text
lost-found-school/
├── main.go, go.mod, go.sum, .env
├── config/ (config.go, database.go)
├── models/ (user, category, report, match, claim, notification, activity_log, settings)
├── controllers/ (auth, report, category, claim, match, notification, admin, page)
├── services/ (matching_service.go, notification_service.go)
├── middleware/ (auth.go, admin.go)
├── routes/routes.go
├── utils/ (response, password, jwt, upload)
├── templates/ (*.html publik, user, admin-*, error)
├── static/ (css/style.css, js/app.js, js/search.js, js/dashboard.js, images/)
├── uploads/  seed/seed.go  README.md
```

Uji dengan Postman/Thunder Client: login → copy `data.token` → header
`Authorization: Bearer <token>`. Browser memakai cookie otomatis.

## Screenshot

Jalankan server dan buka `http://localhost:8080` (tambahkan screenshot di sini).

## Testing Manual (terverifikasi)

27 skenario API lolos: register + validasi + duplikat, login salah, CRUD laporan +
ownership 403, search/filter, klaim + blokir klaim sendiri, verify/reject + alasan wajib,
matching run, approve → RETURNED, notifikasi, dashboard admin, 403 non-admin,
401 tanpa token, activity log, kategori CRUD.

## Future Development

QR Code, notifikasi email/WhatsApp, realtime, mobile app (Flutter), AI image matching,
SSO sekolah, export PDF/Excel, multi-school, cloud storage — arsitektur service-based
siap dikembangkan.
