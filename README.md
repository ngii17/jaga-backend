# JAGA Backend (Working Title)

Platform deteksi & perlindungan dari pinjaman online ilegal dan judi online.
Dibangun untuk tema hackathon **"NextGen Secure: Building the Future of Trusted Web Ecosystems"**.

## Tech Stack

- **Backend:** Golang + Fiber (REST API) + GORM (ORM)
- **Database:** PostgreSQL (dengan ekstensi `pg_trgm` untuk fuzzy search)
- **Autentikasi:** JWT + OTP Email (two-factor)
- **Arsitektur:** Routes → Handler → Service → Repository → Model, dengan interface di layer Repository & Service

## Struktur Project

```
jaga-backend/
├── cmd/api/main.go
├── internal/
│   ├── config/         # baca .env
│   ├── database/       # koneksi PostgreSQL + AutoMigrate + ekstensi pg_trgm
│   ├── models/         # struct GORM untuk semua tabel
│   ├── repository/     # akses data mentah per tabel
│   ├── service/        # logic bisnis per modul
│   ├── handler/        # terima request HTTP
│   ├── middleware/      # JWT auth, permission check, rate limiter
│   └── utils/           # password hash, OTP, JWT, mailer
└── storage/evidence/    # file bukti laporan (terenkripsi, tidak ikut ke Git)
```

## Setup (Windows)

1. Install [Go](https://go.dev/dl) dan [PostgreSQL](https://www.postgresql.org/download/windows/)
2. Buat database baru: `CREATE DATABASE jaga_db;`
3. `copy .env.example .env`, isi sesuai environment kamu (lihat komentar di dalam file)
4. Untuk `SMTP_APP_PASSWORD`, buat lewat [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords) (butuh 2-Step Verification aktif dulu)
5. `go mod tidy`
6. `go run ./cmd/api`

Tabel & ekstensi database akan otomatis dibuat lewat AutoMigrate saat aplikasi pertama kali jalan.

---

## Dokumentasi API

Semua response mengikuti format standar:
```json
{ "success": true, "message": "...", "data": { ... } }
```

Endpoint yang butuh login mengharuskan header:
```
Authorization: Bearer <token>
```

### Modul 1 — Auth

| Method | Endpoint | Auth | Keterangan |
|---|---|---|---|
| POST | `/auth/login` | Publik (rate limited) | Tahap 1: email + password → kirim OTP ke email |
| POST | `/auth/verify-otp` | Publik (rate limited) | Tahap 2: email + kode OTP → dapat token JWT |
| GET | `/auth/me` | Bearer token | Info akun yang sedang login |
| POST | `/auth/users` | Bearer token + `can_manage_users` | Admin mendaftarkan akun verifikator baru |

**Contoh — Login (tahap 1):**
```
POST /auth/login
{ "email": "admin@jaga.id", "password": "rahasia123" }
```

**Contoh — Verifikasi OTP (tahap 2):**
```
POST /auth/verify-otp
{ "email": "admin@jaga.id", "otp": "123456" }
```
Response berisi `data.token` (JWT, berlaku 8 jam) dan `data.user`.

**Contoh — Daftarkan verifikator (khusus admin):**
```
POST /auth/users
Authorization: Bearer <token_admin>
{ "name": "Budi", "email": "budi@gmail.com", "password": "passwordbudi123" }
```

---

### Modul 2 — Cek

Semua endpoint publik (tanpa login), dibatasi rate limiter untuk mencegah penyalahgunaan bot.

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/check/search?q=namaAplikasi` | Cari nama aplikasi/situs — fuzzy search ke data legal & data bahaya |
| GET | `/check/quiz?scenario=pencegahan` | Ambil daftar pertanyaan kuis (`pencegahan` atau `penanganan`) |
| POST | `/check/quiz/submit` | Kirim jawaban kuis, dapat skor & level risiko |

**Contoh — Pencarian:**
```
GET /check/search?q=DanaCepatt
```
Response `data.status` berisi salah satu: `aman`, `mirip_waspada`, `terkonfirmasi_bahaya`.

**Contoh — Submit kuis:**
```
POST /check/quiz/submit
{
  "scenario": "pencegahan",
  "answers": [
    { "question_id": 1, "answer": true },
    { "question_id": 2, "answer": false }
  ]
}
```
Response berisi `data.total_score` dan `data.level_name` (Aman/Waspada/Bahaya).

---

### Modul 3 — Lapor

| Method | Endpoint | Auth | Keterangan |
|---|---|---|---|
| POST | `/report/submit` | Publik (rate limited) | Submit laporan baru (multipart/form-data, bisa sertakan bukti) |
| GET | `/report/:id/status` | Publik | Cek status laporan pakai ID — dipakai pelapor anonim tanpa akun |
| GET | `/report/` | Bearer token + `can_approve_report` | Daftar semua laporan (bisa filter `?status=pending`) |
| GET | `/report/evidence/:evidenceId` | Bearer token + `can_approve_report` | Ambil file bukti yang sudah didekripsi |
| GET | `/report/:id/draft` | Bearer token + `can_approve_report` | Draf teks laporan siap-kirim ke OJK/Satgas PASTI |
| POST | `/report/:id/verify` | Bearer token + `can_approve_report` | Terima/tolak laporan — memicu update `threat_entities` & audit trail |

**Contoh — Submit laporan (form-data, bukan JSON):**
```
POST /report/submit
Content-Type: multipart/form-data

category: pinjol_ilegal
reported_name: DanaCepatt Ilegal
reported_account_or_wa: 0812xxxxxxx
description: Kronologi kejadian...
severity: sudah_diteror
is_anonymous: true
evidence: [file upload]
```
Response berisi `data.report_id` — disimpan pelapor untuk cek status nanti.

**Contoh — Cek status (tanpa login):**
```
GET /report/1/status
```

**Contoh — Verifikasi laporan:**
```
POST /report/7/verify
Authorization: Bearer <token_verifikator>
{ "status": "diterima", "note": "Bukti valid, sesuai laporan" }
```

**Keamanan yang diterapkan di modul ini:**
- Data kontak pelapor (email/telepon) disimpan terpisah dari data laporan, hanya ada jika `is_anonymous = false`
- File bukti dienkripsi (AES) sebelum disimpan ke disk, nama file diacak (UUID), tipe file dideteksi dari isi file (magic bytes) bukan ekstensi
- Setiap perubahan status laporan tercatat permanen di `report_status_logs` (audit trail, hanya insert)
- `submitter_fingerprint` (hash IP + User-Agent) mencegah satu sumber menggelembungkan jumlah laporan

---

### Modul 4 — Komunitas

| Method | Endpoint | Auth | Keterangan |
|---|---|---|---|
| GET | `/testimonials` | Publik | Daftar testimoni yang sudah dipublikasikan |
| POST | `/testimonials/:id/upvote` | Publik | Tambah dukungan "saya juga mengalami ini" |

**Contoh:**
```
GET /testimonials
POST /testimonials/3/upvote
```

---

## Environment Variables (`.env`)

Lihat `.env.example` untuk daftar lengkap. Ringkasannya:

| Variabel | Kegunaan |
|---|---|
| `DATABASE_URL` | Koneksi PostgreSQL |
| `JWT_SECRET` | Kunci tanda tangan token login |
| `SMTP_*` | Kredensial Gmail untuk kirim OTP |
| `EVIDENCE_ENCRYPTION_KEY` | Kunci AES untuk enkripsi file bukti laporan |

**Jangan pernah commit file `.env` asli** — hanya `.env.example` yang boleh ikut ke Git.