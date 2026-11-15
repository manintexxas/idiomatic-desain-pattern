# FIXBACK Admin Portal
## Professional Dashboard Architecture, Implementation Blueprint, dan Demo Plan

**Document type:** System Design, Backend Architecture, UI/UX Navigation, Implementation Roadmap  
**Audience:** Engineering Team, Academic Reviewers, AI Course Evaluators  
**Status:** Proposed Blueprint — implementasi bertahap, bukan klaim bahwa seluruh fitur telah tersedia  
**Target:** V1 yang dapat digunakan dan didemonstrasikan, dengan jalur evolusi menuju production-ready  
**Primary architecture:** Go modular monolith, PostgreSQL + PostGIS, background workers, object storage, AI provider adapters  
**Core principle:** Admin-centric AI caching, zero-AI-cost user ingestion, asynchronous background triage, human-controlled operational lifecycle

---

# 0. Cara Membaca Dokumen

Dokumen ini menjadi sumber acuan implementasi FIXBACK Admin Portal. Urutan pengerjaan tidak mengikuti urutan tampilan dashboard, melainkan mengikuti dependency teknis agar fitur tidak dibangun di atas fondasi yang belum aman.

Setiap bagian menggunakan penanda berikut:

- **MUST:** wajib tersedia agar V1 dapat dipercaya dan digunakan.
- **SHOULD:** sangat dianjurkan, tetapi dapat diselesaikan setelah alur utama stabil.
- **LATER:** ekstensi yang tidak boleh menghambat V1.
- **Acceptance criteria:** kondisi yang harus dapat dibuktikan melalui pengujian atau demonstrasi.

Dokumen ini sengaja memilih solusi sederhana yang dapat ditingkatkan. Jangan langsung memecah sistem menjadi microservices hanya demi terlihat scalable. Modular monolith dengan batas modul yang jelas lebih efisien untuk tahap awal.

## 0.1 Asumsi dan Batasan

1. Backend utama menggunakan Go.
2. Admin portal adalah aplikasi web yang berkomunikasi dengan backend melalui HTTP API.
3. PostgreSQL menjadi sumber kebenaran utama; PostGIS digunakan untuk query geospasial.
4. File gambar disimpan di object storage atau media storage terpisah dari tabel utama.
5. AI hanya dipanggil oleh backend worker atau backend application service. Browser dan end-user tidak memegang API key AI.
6. V1 dapat menjalankan API dan worker dari satu codebase serta satu deployment unit yang dipisahkan menjadi dua process mode.
7. Sistem dapat dimulai dengan satu organisasi/fasilitas, tetapi struktur data sebaiknya menyediakan ruang untuk `tenant_id` bila multi-fasilitas menjadi kebutuhan nyata.
8. Nominal biaya, contoh laporan, koordinat, log, dan statistik pada demo adalah fixture/demo data kecuali benar-benar berasal dari runtime.

**Definisi penting:** “Zero-cost user ingestion” berarti tidak ada pemanggilan model AI saat user mengirim laporan. Penyimpanan file, bandwidth, database, dan CPU lokal tetap mempunyai biaya operasional.

---

# 1. Tujuan Produk dan Batas V1

## 1.1 Tujuan Utama

FIXBACK menerima laporan kerusakan fasilitas, menyimpan bukti dan lokasi, mengelompokkan laporan yang berpotensi duplikat, memproses triage AI secara asynchronous, menyediakan analisis teknis dan draft SPK yang di-cache, lalu mengelola penyelesaian dan pengarsipan laporan.

Admin adalah operator utama sistem. AI berfungsi sebagai decision-support, bukan aktor yang mempunyai wewenang untuk menetapkan penyelesaian pekerjaan atau mengirim instruksi teknis tanpa validasi manusia.

## 1.2 Hasil yang Harus Bisa Dibuktikan

| Kapabilitas | Bukti teknis yang harus ditampilkan |
|---|---|
| Zero-AI-cost ingestion | Request submission tidak memanggil AI provider dan langsung mengembalikan acknowledgement |
| Async triage | Laporan tersimpan sebelum triage selesai; worker memproses job terpisah |
| Cost control | Triage dan deep analysis mempunyai batas pemanggilan, retry, timeout, serta cost tracking |
| Duplicate detection | PostGIS memperkecil kandidat spasial; pHash menghitung kemiripan visual; hasil dicatat dan dapat diaudit |
| Admin-centric analysis | Hanya role admin berizin yang dapat meminta deep analysis |
| Cache lock | Membuka analisis tersimpan tidak memanggil provider AI lagi |
| Lifecycle | Perubahan status mengikuti state transition yang tervalidasi |
| H+1 archiving | Laporan completed diarsipkan setelah 24 jam penuh berdasarkan `completed_at` |
| Observability | Dashboard mengambil metrik dari database, worker heartbeat, dan event log yang benar-benar terjadi |
| Extensibility | Provider AI, prompt, threshold duplicate, dan job handler dapat ditambah tanpa mengubah seluruh alur |

## 1.3 Non-Goals V1

Jangan memasukkan hal berikut sebelum alur dasar stabil:

- Microservices dan distributed message broker yang belum dibutuhkan.
- Real-time websocket untuk setiap perubahan kecil; polling terukur atau Server-Sent Events cukup sebagai awal.
- Multi-agent AI yang saling memanggil tanpa kebutuhan terukur.
- AI yang otomatis menyetujui pekerjaan konstruksi atau mengirim SPK final.
- Automatic payment, procurement, atau accounting.
- Integrasi WhatsApp otomatis sebelum kanal resmi, consent, template pesan, dan audit pengiriman tersedia.
- Model computer vision khusus yang dilatih sendiri sebelum dataset dan baseline tersedia.

---

# 2. Keputusan Arsitektur Inti

## 2.1 Bentuk Arsitektur

Gunakan **modular monolith**: satu codebase dan satu domain model, tetapi setiap modul memiliki tanggung jawab dan interface internal yang jelas.

```text
                          ┌──────────────────────────┐
                          │      Web Applications   │
                          │  User App / Admin Portal│
                          └────────────┬─────────────┘
                                       │ HTTPS / JSON
                          ┌────────────▼─────────────┐
                          │       Go HTTP API        │
                          │ Auth / Validation / DTO  │
                          └────────────┬─────────────┘
                                       │
                       ┌───────────────▼────────────────┐
                       │       Application Layer       │
                       │ Reports / Triage / Analysis    │
                       │ Clustering / Lifecycle / Audit │
                       └───────┬───────────┬────────────┘
                               │           │
                 ┌─────────────▼──┐   ┌────▼──────────────────┐
                 │ PostgreSQL     │   │ Object / Media Store  │
                 │ + PostGIS      │   │ Image Originals       │
                 │ Jobs + Records │   │ Thumbnails             │
                 └─────────────┬──┘   └───────────────────────┘
                               │
                       ┌───────▼────────┐
                       │ Go Worker Mode │
                       │ Job Poller     │
                       │ Retry / Lease  │
                       └───────┬────────┘
                               │
                 ┌─────────────▼──────────────────┐
                 │ Provider Adapter Layer         │
                 │ Triage AI / Deep Analysis AI   │
                 │ Local pHash / Other AI Provider│
                 └────────────────────────────────┘
```

API process dan worker process boleh berjalan sebagai deployment terpisah, tetapi memakai package domain/application yang sama. Dengan demikian, deployment dapat ditambah tanpa menggandakan business logic.

## 2.2 Aturan Boundary Modul

| Modul | Bertanggung jawab | Tidak boleh bertanggung jawab |
|---|---|---|
| Identity & Access | Login, session/token, role, permission | Business status report |
| Report Intake | Validasi submission, metadata, media, idempotency | Memanggil AI provider |
| Duplicate Detection | Spatial candidate, pHash, hubungan duplicate | Menghapus bukti submission |
| Job Queue | Claim, lease, retry, dead-letter, job history | Menyimpan business decision AI sebagai satu-satunya sumber data |
| AI Triage | Summary, category, baseline severity, confidence | Mengubah report menjadi completed |
| Deep Analysis | Technical proposal, budget estimate, SPK draft, cache/version lock | Mengirim instruksi kerja final tanpa approval |
| Lifecycle | State transition dan archive policy | Menghasilkan analisis AI |
| Admin Query | Dashboard aggregation, filter, pagination | Mengubah data domain secara langsung |
| Audit & Observability | Audit event, worker heartbeat, metrics | Menjadi sumber kebenaran status report |

## 2.3 Aturan Dependency

- HTTP handler hanya memanggil application service.
- Application service menjalankan use case dan mengatur transaksi.
- Repository mengabstraksikan akses database.
- AI provider adapter hanya dipanggil dari service/worker yang berwenang.
- Worker tidak boleh menulis tabel domain dengan aturan berbeda dari application service. Worker menggunakan service yang sama atau domain operation yang sama.
- Frontend tidak boleh menghitung keputusan bisnis seperti duplicate threshold atau status transition.
- Database constraints tetap diperlukan karena validasi aplikasi saja tidak cukup untuk mencegah race condition.

## 2.4 Sumber Kebenaran

- PostgreSQL: status report, relasi duplicate, votes, AI result, analysis version, job status, audit.
- Object storage: binary image dan dokumen export.
- AI provider: hanya menghasilkan output sementara; hasil yang diterima dan divalidasi harus disimpan sebelum dianggap sebagai hasil sistem.
- Frontend state: cache tampilan sementara, bukan sumber kebenaran.
- Worker log: observability, bukan sumber status domain.

---

# 3. Domain Model dan State Machine

Salah satu risiko desain awal adalah mencampurkan status report dengan status AI. Ketiganya wajib dipisahkan.

## 3.1 Tiga Jenis Status

### A. Report Lifecycle Status

Status ini menunjukkan perjalanan operasional sebuah laporan.

| Status | Makna | Masuk dari | Keluar ke |
|---|---|---|---|
| `submitted` | Laporan tersimpan dan menunggu triage/review | Intake berhasil | `in_review`, `duplicate` |
| `in_review` | Laporan sedang dinilai admin atau menunggu keputusan | Triage berhasil atau admin override | `processing`, `completed` bila memang tidak memerlukan pekerjaan |
| `processing` | Pekerjaan perbaikan sedang dilakukan | Admin memulai pekerjaan | `completed` |
| `completed` | Pekerjaan dinyatakan selesai oleh pihak berwenang | Admin/teknisi dengan permission | `archived` setelah 24 jam |
| `archived` | Laporan read-only berdasarkan policy H+1 | Archiver | Tidak ada transisi normal |
| `duplicate` | Submission ini ditautkan ke report canonical lain | Duplicate service/admin | Unlink/reclassify melalui tindakan admin yang diaudit |

Status `duplicate` adalah status operasional untuk submission yang bukan master. Record submission tetap dipertahankan.

### B. Triage Status

| Status | Makna |
|---|---|
| `pending` | Menunggu job triage |
| `processing` | Job sedang dikerjakan worker |
| `succeeded` | Output AI berhasil divalidasi dan disimpan |
| `failed` | Pemrosesan gagal setelah aturan retry atau ditandai gagal permanen |
| `skipped` | Tidak diproses karena duplicate, file tidak valid, atau kebijakan lain |

### C. Deep Analysis Status

| Status | Makna |
|---|---|
| `not_requested` | Belum ada permintaan analisis mendalam |
| `pending` | Permintaan sudah tercatat dan menunggu worker |
| `generating` | Provider AI sedang dipanggil |
| `locked` | Hasil final versi tersebut sudah disimpan dan dikunci |
| `failed` | Analisis gagal; hasil final belum tersedia |

**Penting:** `locked` berarti hasil analisis tersimpan sebagai snapshot immutable. Cache hit berarti sistem menggunakan snapshot tersimpan tanpa memanggil AI.

## 3.2 State Transition Rules

Semua perubahan status melewati Lifecycle Service, bukan update bebas dari handler atau frontend.

| Dari | Ke | Aktor | Validasi wajib |
|---|---|---|---|
| submitted | in_review | Worker atau Admin | Triage succeeded atau admin melakukan override dengan alasan |
| submitted | duplicate | Duplicate Service | Master report valid, bukan archived/duplicate, relasi lolos aturan |
| in_review | processing | Admin/Technician role | Aktor memiliki permission dan assignment bila dipakai |
| in_review | completed | Admin dengan permission khusus | Alasan/hasil penyelesaian tercatat; hanya untuk pekerjaan yang tidak membutuhkan status processing |
| processing | completed | Admin/Technician role | Bukti atau catatan penyelesaian sesuai policy |
| completed | archived | Archiver | `completed_at` lebih lama atau sama dengan cutoff 24 jam |
| archived | status lain | Tidak diizinkan secara normal | Reopen khusus, alasan wajib, audit event wajib |

Aturan tambahan:

1. `completed_at` hanya diisi ketika transisi menuju completed.
2. Update catatan, upvote, atau metadata tidak boleh mengubah `completed_at`.
3. `archived_at` hanya diisi ketika archiver berhasil melakukan transisi.
4. Duplicate report tidak boleh ikut diproses sebagai report canonical.
5. Perubahan status harus menggunakan compare-and-set atau row lock agar dua admin tidak menimpa perubahan bersamaan.
6. Reopen archived bukan operasi biasa. V1 dapat menolak reopen; apabila diperlukan, sediakan administrative override yang tercatat.

## 3.3 Workflow Utama

```text
User Submission
      │
      ▼
Report Stored (status=submitted, triage=pending)
      │
      ▼
Dedupe Preflight (local pHash + PostGIS candidate lookup)
      │
      ├── Strong Match ──► Relation to Master + status=duplicate
      │                         │
      │                         └── Idempotent auto-upvote on Master
      │
      ├── Partial Match ─► Flag candidate for admin review
      │
      └── No Match ──────► Enqueue Triage Job
                                  │
                                  ▼
                             AI Triage Worker
                                  │
                       ┌──────────┴──────────┐
                       ▼                     ▼
                   Success                 Failure
                       │                     │
             Save structured result     Retry / failed state
                       │
                       ▼
              report.status=in_review
                       │
              Admin requests Deep Analysis
                       │
                       ▼
              Deep Analysis Job + Cache Guard
                       │
                       ▼
              Store immutable analysis version
                       │
                       ▼
                 Admin Review / SPK
                       │
                       ▼
              in_review → processing
                       │
                       ▼
                  completed
                       │
                  Wait 24 hours
                       │
                       ▼
                   archived
```

---

# 4. Data Model dan Database Design

Gunakan UUID atau identifier yang tidak mudah ditebak untuk primary key internal. Public reference seperti `FIX-2026-00024` boleh dipakai untuk tampilan, tetapi jangan menjadikannya satu-satunya primary key.

## 4.1 Tabel Minimum V1

### `users`

| Field | Tipe konseptual | Keterangan |
|---|---|---|
| id | UUID | Primary key |
| email | string | Unique, normalized |
| password_hash / external_subject | string | Sesuai strategi authentication |
| display_name | string | Nama tampilan |
| is_active | boolean | Akun dapat dinonaktifkan |
| created_at, updated_at | timestamp UTC | Audit waktu |

### `roles` dan `user_roles`

Role awal: `admin`, `facility_manager`, `technician`, `reviewer` (pilih hanya yang benar-benar digunakan V1). Gunakan permission yang eksplisit untuk aksi sensitif seperti `analysis.generate`, `report.transition`, `report.archive.override`, dan `spk.approve`.

### `reports`

| Field | Tipe konseptual | Keterangan |
|---|---|---|
| id | UUID | Primary key |
| public_ref | string | Nomor yang dibaca manusia, unique |
| tenant_id | UUID nullable | Disiapkan bila multi-fasilitas diperlukan |
| reporter_id | UUID nullable | Boleh null untuk laporan anonim sesuai policy |
| title | string | Judul singkat |
| description | text | Deskripsi dari pelapor |
| category | string/enum | Kategori awal, boleh unknown |
| location | geography(Point, 4326) | Koordinat untuk PostGIS |
| location_accuracy_m | numeric nullable | Akurasi GPS jika tersedia |
| status | enum/string constrained | Report lifecycle status |
| triage_status | enum/string constrained | Triage state |
| severity | enum nullable | Baseline severity dari triage/admin |
| triage_summary | text nullable | Summary hasil triage yang telah divalidasi |
| completed_at | timestamp nullable | Waktu laporan dinyatakan selesai |
| archived_at | timestamp nullable | Waktu diarsipkan |
| created_at, updated_at | timestamp UTC | Metadata waktu |

**Catatan:** hindari menaruh seluruh data AI ke satu kolom `reports`. Tabel report menyimpan ringkasan operasional yang sering dibaca. Detail run AI disimpan di tabel tersendiri.

### `report_media`

| Field | Tipe konseptual | Keterangan |
|---|---|---|
| id | UUID | Primary key |
| report_id | UUID | Foreign key |
| object_key | string | Key/path object storage, bukan public URL permanen |
| media_type | string | MIME yang telah diverifikasi dari konten |
| byte_size | integer | Ukuran file |
| checksum_sha256 | string | Deteksi integritas/duplikasi file identik |
| phash_value | string/binary | Perceptual hash |
| phash_algorithm_version | string | Versi algoritma agar hash yang tidak kompatibel tidak dibandingkan |
| width, height | integer | Dimensi gambar setelah decode |
| created_at | timestamp UTC | Waktu penyimpanan |

### `triage_runs`

Menyimpan histori setiap percobaan triage.

Field minimum: `id`, `report_id`, `job_id`, `provider`, `model`, `prompt_version`, `status`, `input_fingerprint`, `summary`, `category`, `severity`, `confidence`, `raw_response_reference` (opsional dan dibatasi), `error_code`, `started_at`, `finished_at`, `created_at`.

Raw response tidak harus disimpan seluruhnya bila mengandung informasi sensitif atau biaya penyimpanan tinggi. Simpan hasil terstruktur yang telah divalidasi.

### `report_relations`

Menyimpan hubungan duplicate/candidate tanpa menghapus report.

Field minimum: `id`, `source_report_id`, `target_report_id`, `relation_type`, `match_reason`, `spatial_distance_m`, `phash_distance`, `confidence_or_score` (nullable), `status`, `created_by`, `created_at`, `reviewed_by`, `reviewed_at`.

`relation_type` awal: `duplicate_of`, `possible_duplicate_of`. `status`: `proposed`, `confirmed`, `rejected`, `reversed`.

### `report_votes`

Menyimpan upvote manusia maupun otomatis secara idempotent.

Field minimum: `id`, `report_id`, `actor_type`, `actor_id` nullable, `source_report_id` nullable, `idempotency_key`, `created_at`.

Constraint penting: `idempotency_key` unik. Untuk human vote, kombinasi report dan user juga harus unik sesuai policy. Jangan hanya menambah integer `upvotes` tanpa menyimpan sumber event; retry worker dapat menggandakan angka.

### `analyses`

Header untuk sebuah analysis version.

Field minimum: `id`, `report_id`, `version_number`, `status`, `input_fingerprint`, `provider`, `model`, `prompt_version`, `generated_by`, `locked_by`, `locked_at`, `generated_at`, `created_at`, `failure_code`.

Unique constraint: kombinasi report dan version number. Hanya satu request aktif untuk fingerprint yang sama pada satu report.

### `analysis_contents`

Simpan output terstruktur, misalnya:

- technical assessment;
- recommended restoration steps;
- risk and safety notes;
- estimated material list;
- labor estimate;
- total budget range;
- assumptions and exclusions;
- SPK draft fields;
- AI disclaimer;
- human approval metadata.

Pisahkan isi dari metadata agar payload dapat berkembang tanpa membuat tabel utama penuh kolom yang jarang dipakai.

### `jobs`

| Field | Keterangan |
|---|---|
| id | UUID |
| job_type | Contoh: dedupe, triage, deep_analysis, archive_sweep |
| aggregate_id | ID report atau entity terkait |
| status | queued, running, succeeded, retry_wait, dead_letter, cancelled |
| payload | Parameter non-sensitive yang diperlukan handler |
| idempotency_key | Unique per logical job |
| attempts / max_attempts | Jumlah percobaan |
| available_at | Waktu paling awal boleh diproses |
| locked_by | Worker pemilik lease |
| locked_until | Batas lease |
| last_error_code | Kode error terklasifikasi |
| created_at / updated_at | Metadata |

Payload job jangan menjadi tempat menyimpan binary image atau API secret. Gunakan object key/reference.

### `audit_logs`

Field minimum: `id`, `actor_type`, `actor_id`, `action`, `entity_type`, `entity_id`, `before_snapshot` (terbatas), `after_snapshot` (terbatas), `reason`, `request_id`, `ip_hash_or_ip` sesuai kebijakan privasi, `created_at`.

Audit log bersifat append-only dari aplikasi. Jangan mengandalkan log console sebagai audit perubahan bisnis.

### `worker_heartbeats`

Field minimum: `worker_id`, `worker_type`, `instance_id`, `last_seen_at`, `started_at`, `version`, `metadata` terbatas.

Heartbeat tidak membuktikan sebuah job sehat; gabungkan dengan job throughput, error, dan stuck lease.

## 4.2 Indexes Wajib

1. `reports(status, created_at)` untuk queue dan active listing.
2. `reports(triage_status, created_at)` bila queue triage difilter dari report.
3. GiST index pada `reports.location` untuk PostGIS.
4. `report_media(report_id)`.
5. `report_relations(source_report_id)` dan `report_relations(target_report_id)`.
6. `jobs(status, available_at)` untuk claim job.
7. `jobs(locked_until)` untuk mendeteksi lease kedaluwarsa.
8. `analyses(report_id, version_number)`.
9. `audit_logs(entity_type, entity_id, created_at)`.
10. Unique index pada public reference, idempotency key, dan constraint vote.

Jangan menambah index di semua kolom. Index tambahan harus berdasarkan query nyata dan `EXPLAIN ANALYZE`.

## 4.3 Database Constraints dan Integrity

- Foreign key untuk hubungan yang wajib konsisten.
- Check constraint untuk nilai status yang valid jika enum belum digunakan.
- Unique idempotency key untuk request/job/vote.
- Report tidak boleh menunjuk dirinya sendiri sebagai duplicate.
- Hubungan duplicate tidak boleh membentuk cycle.
- Master duplicate harus valid dan bukan archived/duplicate, kecuali policy khusus mengizinkan.
- `completed_at` wajib terisi ketika status completed/archived.
- `archived_at` wajib terisi ketika status archived.
- Data analisis locked tidak boleh diubah in-place.

Constraint yang tidak dapat diekspresikan dengan mudah di database harus dijaga melalui transaction dan domain service.

---

# 5. Report Intake — Zero-AI-Cost Ingestion

## 5.1 Target Perilaku

User dapat mengirim laporan tanpa menunggu AI. Backend menyimpan data yang valid, mengembalikan response acknowledgement, dan memasukkan pekerjaan asynchronous ke job queue.

**Tidak boleh ada** pemanggilan Gemini/VLM/LLM dari jalur HTTP intake.

## 5.2 Urutan Intake

1. Terima request dan validasi authentication/anonymous policy.
2. Validasi ukuran request, field wajib, koordinat, dan format media.
3. Terapkan rate limit per user/IP/device sesuai risiko abuse.
4. Validasi idempotency key agar request retry dari jaringan buruk tidak membuat report baru.
5. Decode/inspeksi media secara terbatas: MIME sniffing, batas dimensi, ukuran, dan file signature.
6. Simpan file ke object storage dengan object key internal.
7. Simpan report dan media metadata ke database.
8. Buat job dedupe atau preprocessing sesuai pipeline.
9. Commit transaksi database.
10. Kembalikan acknowledgement dengan report ID/reference dan status awal.

Jika object storage berhasil tetapi database gagal, file menjadi orphan object. Sediakan cleanup job untuk object yang tidak direferensikan setelah grace period. Jika database berhasil tetapi upload gagal, jangan membuat report yang seolah mempunyai bukti lengkap.

## 5.3 Upload Strategy

### Pilihan V1

Untuk file kecil dan beban awal rendah, backend dapat menerima multipart upload. Batas ukuran harus eksplisit dan configurable.

### Jalur Skalabilitas

Pindah ke two-step upload:

1. Client meminta upload authorization.
2. Backend membuat object key dan signed upload URL berumur pendek.
3. Client upload langsung ke object storage.
4. Client mengirim finalize request.
5. Backend memverifikasi object benar-benar ada dan cocok dengan ukuran/metadata.
6. Backend membuat report dan job.

Jangan memberikan public permanent URL untuk media laporan. Gunakan signed read URL berumur pendek ketika admin berhak melihat gambar.

## 5.4 Validasi Media dan Privasi

- Periksa file berdasarkan konten, bukan hanya ekstensi.
- Batasi ukuran file dan dimensi gambar.
- Tolak format yang tidak didukung.
- Pertimbangkan decompression bomb dan file malformed.
- Decode memakai resource limit/time limit.
- Strip EXIF dari derivative/thumbnail untuk mencegah kebocoran metadata perangkat; simpan koordinat lokasi yang memang dikirim melalui field terpisah dan sesuai consent.
- Jangan menulis binary, signed URL, atau informasi sensitif ke log.
- Thumbnail dibuat asynchronous bila pembuatan thumbnail menambah latency intake secara berarti.

## 5.5 Acceptance Criteria Intake

- Submission valid tersimpan walaupun AI provider tidak tersedia.
- Tidak ada AI call pada request intake, dibuktikan menggunakan fake provider/call counter.
- Request yang dikirim ulang dengan idempotency key sama tidak membuat report kedua.
- File terlalu besar, MIME palsu, koordinat invalid, dan payload malformed ditolak dengan error konsisten.
- Kegagalan object storage/database tidak meninggalkan report setengah valid.
- Public API tidak pernah mengembalikan API key, raw internal object key, atau data admin-only.

---

# 6. Duplicate Detection — PostGIS + pHash

## 6.1 Tujuan dan Prinsip

Duplicate detection mengurangi ticket ganda dan pemanggilan AI yang tidak perlu. Proses ini harus murah, explainable, dan dapat dikoreksi admin.

Jangan menganggap satu sinyal similarity sebagai bukti pasti bahwa dua laporan adalah masalah yang sama.

## 6.2 Dua Tahap Pencarian

### Tahap A — Spatial Candidate Search

- Gunakan PostGIS `ST_DWithin` pada tipe geography agar radius dinyatakan dalam meter.
- Radius awal mengikuti rancangan demo: 15 meter.
- Kandidat harus dibatasi pada report aktif dan canonical yang relevan.
- Jangan mengambil seluruh tabel report untuk dibandingkan di memory.
- Gunakan GiST spatial index.
- Radius harus configurable, bukan angka yang tersebar di banyak file.

### Tahap B — Visual Similarity

- pHash adalah perceptual hash yang merepresentasikan karakter visual gambar menjadi hash ringkas.
- Untuk hash 64-bit dalam format hex 16 karakter, hitung Hamming Distance melalui XOR lalu menghitung jumlah bit yang berbeda.
- Semakin rendah jarak, semakin mirip representasi visualnya.
- Threshold awal demo: `distance <= 5`, dengan syarat kedua hash berasal dari algoritma/version yang kompatibel.
- Simpan nilai jarak dan versi algoritma agar keputusan dapat diaudit.

## 6.3 Default Decision Policy

Gunakan kebijakan konservatif berikut untuk V1:

| Spatial match | pHash match | Hasil default |
|---|---|---|
| Ya | Ya | Strong duplicate candidate; boleh auto-link jika report canonical masih aktif dan aturan tambahan lolos |
| Ya | Tidak | Possible duplicate; tampilkan untuk review atau biarkan menjadi report terpisah |
| Tidak | Ya | Possible duplicate; kemungkinan gambar yang sama di lokasi berbeda, jangan auto-merge |
| Tidak | Tidak | Tidak dianggap duplicate |

Jika sistem mempunyai kategori yang dapat dipercaya, gunakan kompatibilitas kategori sebagai sinyal tambahan. Jangan memakai OR rule untuk auto-merge karena dapat menggabungkan laporan yang lokasinya sama tetapi kerusakannya berbeda, atau gambar sama yang dipakai ulang pada lokasi berbeda.

Threshold radius 15m dan Hamming Distance 5 adalah **initial configuration**, bukan angka universal. Kalibrasikan dengan dataset nyata.

## 6.4 Simpan Duplicate Submission, Jangan Hilangkan Evidence

Ketika report baru dinyatakan duplicate:

1. Pertahankan row report baru.
2. Pertahankan media dan waktu submission sesuai retention policy.
3. Buat relasi `duplicate_of` ke master report.
4. Ubah status submission menjadi `duplicate`.
5. Buat auto-vote/upvote event dengan idempotency key yang berasal dari source report.
6. Tandai triage sebagai skipped jika tidak perlu memanggil AI untuk duplicate.
7. Tampilkan hubungan ini pada cluster detail.

Dengan desain ini, sistem tidak membuat ticket aktif kedua, tetapi tidak menghilangkan bukti bahwa ada orang lain yang melaporkan masalah tersebut.

## 6.5 Race Condition pada Ingestion Bersamaan

Dua laporan serupa dapat tiba hampir bersamaan. Jika dua worker mencari kandidat sebelum salah satunya menjadi canonical, keduanya dapat lolos sebagai master.

Mitigasi V1:

- Serialize dedupe untuk wilayah/tile yang sama menggunakan transaction-scoped advisory lock atau mekanisme lock ekuivalen.
- Bila menggunakan tile/geohash, pertimbangkan tile tetangga karena radius 15m dapat melintasi batas tile.
- Ambil lock dalam urutan deterministik untuk menghindari deadlock.
- Setelah lock diperoleh, jalankan ulang spatial candidate query di dalam transaksi.
- Gunakan unique/idempotency constraints untuk auto-vote dan relation creation.
- Bila skala masih kecil, worker concurrency dedupe dapat dibatasi sebagai langkah awal, tetapi tetap siapkan database integrity.

Jangan mengandalkan pemeriksaan di frontend atau query tanpa lock untuk mencegah duplicate race.

## 6.6 Koreksi False Positive

Admin harus dapat:

- melihat alasan keputusan: jarak spasial, Hamming Distance, waktu, kategori;
- menolak possible duplicate;
- memutus relasi duplicate yang salah;
- mengembalikan report ke alur review bila sebelumnya salah digabung;
- melihat audit siapa yang mengoreksi dan alasannya.

Auto-upvote yang berasal dari duplicate relation juga harus dapat direkonsiliasi ketika relation dibatalkan. Jangan menghapus human vote ketika auto-vote dikoreksi.

## 6.7 Acceptance Criteria Duplicate

- Query spatial memakai index yang sesuai.
- Hash yang tidak kompatibel tidak dibandingkan secara diam-diam.
- Match dua sinyal menghasilkan relasi sesuai policy.
- Match satu sinyal tidak auto-merge.
- Dua worker concurrent tidak membuat dua master untuk kasus yang sama tanpa terdeteksi.
- Retry job tidak menambah upvote berulang.
- Admin dapat melihat dan membatalkan relasi duplicate.

---

# 7. Job Queue dan Background Worker

## 7.1 Pilihan Queue V1

Gunakan tabel `jobs` di PostgreSQL sebagai database-backed queue. Ini mengurangi infrastruktur dan cukup untuk beban awal.

Jangan mengandalkan in-memory channel sebagai satu-satunya queue, karena job akan hilang ketika process restart.

Jika throughput atau pola distribusi meningkat, queue dapat dipindahkan ke broker eksternal tanpa mengubah domain use case.

## 7.2 Aturan Delivery

Anggap delivery bersifat **at-least-once**: job yang sama dapat dijalankan lebih dari satu kali karena process crash, lease timeout, atau retry.

Akibatnya:

- setiap handler wajib idempotent;
- side effect harus mempunyai idempotency key;
- jangan berasumsi bahwa satu job pasti hanya dieksekusi sekali;
- job sukses harus tersimpan secara durable sebelum worker menganggap proses selesai.

## 7.3 Job Claim dan Lease

1. Worker mencari job `queued` atau `retry_wait` dengan `available_at` yang sudah tiba.
2. Worker claim job secara atomik; gunakan row locking seperti `SKIP LOCKED` atau mekanisme ekuivalen.
3. Worker mengisi `locked_by` dan `locked_until`.
4. Worker memperbarui lease secara berkala untuk pekerjaan panjang.
5. Setelah sukses, worker menandai `succeeded`.
6. Jika process mati dan lease kedaluwarsa, job dapat di-claim ulang.
7. Retry meningkatkan attempt count dan mengatur `available_at` berdasarkan backoff.

## 7.4 Retry Policy

- Retry hanya untuk error transient: timeout, temporary network failure, rate limit, provider 5xx, database connection interruption.
- Jangan retry error permanent tanpa perubahan input: invalid image, schema invalid berulang, unauthorized provider, unsupported media.
- Gunakan exponential backoff dengan jitter.
- Hormati `Retry-After` bila provider mengirimnya.
- Batasi `max_attempts` per job type.
- Setelah batas retry, pindahkan ke `dead_letter` dan tampilkan alasan pada admin hub.
- Sediakan manual retry dengan audit trail.

## 7.5 Worker Concurrency dan Resource Control

Konfigurasi terpisah:

- jumlah worker per job type;
- maximum batch size;
- maximum simultaneous AI calls;
- provider rate limit;
- request timeout;
- job lease duration;
- retry limit;
- memory limit untuk image decoding.

Jangan menjalankan image decode, hashing, dan AI call dalam jumlah tidak terbatas. Batasi concurrency agar satu burst upload tidak menghabiskan memory atau quota provider.

## 7.6 Batch Semantics

Batch worker adalah pengelompokan job untuk efisiensi, bukan satu transaksi besar.

- Setiap report mempunyai hasil dan status sendiri.
- Kegagalan satu report tidak boleh membatalkan hasil report lain yang telah sukses.
- Batch menyimpan `batch_id`, start/end time, count sukses/gagal/skipped.
- Manual trigger membuat batch request, bukan menjalankan AI langsung di HTTP request.
- Endpoint manual trigger mengembalikan batch/job ID dan UI melakukan polling.

## 7.7 Worker Health

Heartbeat diperbarui berkala. Worker dianggap stale bila `last_seen_at` melewati ambang yang dikonfigurasi.

Tampilkan minimal:

- worker process count;
- heartbeat terakhir;
- active job;
- queued count;
- oldest queued age;
- retry count;
- dead-letter count;
- job success/failure rate;
- average/p95 processing duration.

Jangan menampilkan worker `ONLINE` hanya karena process API hidup.

---

# 8. AI Integration Architecture

## 8.1 Provider Adapter

Buat contract internal terpisah untuk:

- Triage provider: model ringan, output singkat dan terstruktur.
- Deep analysis provider: model lebih kuat untuk technical assessment, material estimate, labor estimate, dan SPK draft.
- Future provider: vendor lain, local model, atau model khusus.

Business service tidak boleh tergantung langsung pada SDK vendor. Provider-specific request/response, retry code, usage metadata, dan error mapping berada di adapter.

## 8.2 Prompt Versioning

Setiap prompt mempunyai:

- prompt ID;
- prompt version;
- tujuan task;
- expected schema;
- model compatibility;
- created_at dan status active/deprecated;
- changelog singkat.

Perubahan prompt dapat mengubah output dan biaya. Simpan versi prompt pada setiap triage run dan analysis version.

## 8.3 Structured Output Validation

AI output harus diperlakukan sebagai data tidak tepercaya.

1. Parse response.
2. Validasi JSON/schema dan required fields.
3. Validasi enum severity/category.
4. Batasi panjang string dan jumlah item.
5. Validasi numeric budget tidak negatif dan berada dalam batas wajar konfigurasi.
6. Tolak field tambahan yang tidak diharapkan bila schema strict digunakan.
7. Simpan output yang lolos validasi.
8. Jika gagal, catat error dan gunakan retry policy yang sesuai.

Jangan mengirim raw AI output langsung ke HTML. Escape/render sebagai text, bukan executable markup.

## 8.4 Triage Output Contract

Output triage minimum:

- summary;
- detected category;
- baseline severity (`low`, `medium`, `high`, `critical`, atau `unknown`);
- confidence score jika provider mendukung dan definisinya jelas;
- evidence/observations yang terlihat;
- uncertainty/limitations;
- prompt version;
- provider/model metadata;
- usage/cost metadata yang tersedia.

Severity AI adalah rekomendasi awal. Admin tetap dapat mengoreksi dengan audit trail.

## 8.5 Deep Analysis Output Contract

Output deep analysis minimum:

1. Ringkasan kondisi.
2. Observasi yang dapat didukung oleh gambar/deskripsi.
3. Tingkat risiko dan ketidakpastian.
4. Rekomendasi pemeriksaan lapangan.
5. Restoration steps sebagai draft.
6. Material estimate beserta quantity/unit/assumption.
7. Labor estimate beserta assumption.
8. Budget estimate atau range, bukan angka presisi palsu.
9. SPK draft: nomor, lokasi, deskripsi pekerjaan, lingkup, tim tujuan, deadline proposal, catatan keselamatan.
10. Disclaimer bahwa estimasi dan metode kerja harus diverifikasi teknisi/engineer berwenang.

Model tidak boleh mengarang harga supplier seolah merupakan harga aktual. Tandai harga sebagai estimasi dan catat sumber/asumsi. Untuk isu struktur, listrik, kebakaran, gas, atau bahaya keselamatan, analisis AI tidak menggantikan inspeksi ahli.

## 8.6 Input Safety dan Prompt Injection

Gambar dan teks laporan adalah input tidak tepercaya. Bisa saja mengandung instruksi yang mencoba mengubah perilaku model.

Mitigasi:

- pisahkan system instruction dari report content;
- nyatakan bahwa teks dalam gambar/deskripsi adalah evidence, bukan instruksi kepada model;
- jangan izinkan model menjalankan tool/action administratif;
- jangan berikan credential atau secret pada prompt;
- batasi data yang dikirim ke provider;
- gunakan timeout dan output token limit;
- simpan prompt version dan model metadata;
- lakukan human review terhadap hasil dengan risiko tinggi.

## 8.7 Cost Governance

Setiap AI call harus mempunyai pencatatan:

- task type;
- report ID;
- provider dan model;
- prompt version;
- input/output token atau usage unit bila tersedia;
- estimated cost;
- execution duration;
- success/failure/unknown outcome;
- request correlation ID.

Policy minimum:

- Triage hanya satu logical run sukses per input fingerprint, kecuali reprocess eksplisit.
- Deep analysis hanya atas aksi admin berizin.
- Deep analysis yang locked tidak dipanggil ulang ketika dibuka.
- Limit jumlah deep analysis per report/version.
- Global daily/monthly budget dan per-task budget dapat dikonfigurasi.
- Jika budget mencapai threshold, hentikan generation baru atau minta override admin berizin.
- Jangan menyimpan API key di frontend, repository, database plaintext, atau log.

**Perhatian timeout:** timeout setelah request dikirim tidak selalu berarti provider tidak memproses request. Jika outcome tidak diketahui, tandai sebagai `unknown`/reconciliation-needed dan jangan langsung melakukan retry mahal tanpa kebijakan. Gunakan provider idempotency key bila didukung.

---

# 9. Deep Analysis, Cache Guard, dan SPK

## 9.1 Prinsip UX

Ketika admin membuka report:

- Jika ada analysis version locked, tampilkan hasil yang tersimpan.
- Jika belum ada, tampilkan state `Not Generated` dan tombol `Generate Deep Analysis`.
- Jangan memicu AI hanya karena halaman report dibuka atau dropdown berubah.
- Sebelum generation, tampilkan konfirmasi bahwa proses menggunakan AI dan dapat mengonsumsi quota.
- Setelah sukses, tampilkan waktu generation, model/prompt version, status lock, dan siapa yang meminta.

Ini sedikit lebih aman daripada otomatis menjalankan AI pada first view. Hasil akhirnya tetap sama: generation pertama menyimpan hasil, pembukaan berikutnya menggunakan cache.

## 9.2 Cache Guard Workflow

1. Admin meminta generation dan permission diverifikasi.
2. Backend membuat analysis request dengan idempotency key.
3. Backend memeriksa analysis version locked dengan input fingerprint yang sama.
4. Jika ditemukan, kembalikan cache hit tanpa provider call.
5. Jika generation untuk fingerprint yang sama sedang berjalan, kembalikan status pending/generating yang sama.
6. Jika tidak ada, buat analysis version pending dan job deep-analysis secara atomik.
7. Worker claim job dan menandai generating.
8. Worker memanggil provider di luar transaksi database panjang.
9. Worker memvalidasi output dan menulis content + metadata.
10. Worker mengunci version: `status=locked`, `locked_at`, `locked_by`/system actor.
11. Admin UI mengambil hasil tersimpan.

Jangan menahan database transaction/row lock selama menunggu AI provider. Buat state `generating` terlebih dahulu, lakukan external call, lalu simpan hasil dalam transaksi baru.

## 9.3 Cegah Double Charge dari Concurrent Request

Dua admin dapat meminta analysis pada report yang sama hampir bersamaan.

Mitigasi:

- idempotency key per logical request;
- unique constraint untuk active analysis fingerprint;
- transaction saat membuat analysis request/job;
- hanya satu job aktif untuk report + fingerprint;
- worker lease dan idempotent completion;
- cache check dilakukan di backend, bukan frontend;
- request kedua mendapat status request pertama, bukan membuat provider call kedua.

## 9.4 Immutability dan Re-Analysis

Jangan mengedit hasil locked secara in-place.

Jika hasil perlu diperbaiki:

- buat analysis version baru;
- simpan version lama;
- wajibkan alasan re-analysis;
- catat aktor yang meminta;
- simpan prompt/model/input fingerprint baru;
- tandai versi mana yang menjadi current version.

Dengan demikian, “locked” berarti immutable per version, bukan sistem tidak pernah boleh diperbarui sepanjang umur produk.

## 9.5 Human Approval untuk SPK

Status yang disarankan:

`draft_generated` → `reviewed` → `approved` → `issued` → `closed`.

V1 minimum dapat menyimpan SPK draft dan approval admin. Jangan menyamakan AI-generated dengan approved/issued.

PDF export harus mengambil data yang sudah disetujui atau secara eksplisit diberi watermark `DRAFT`. Pengiriman WhatsApp otomatis ditunda hingga tersedia provider resmi, consent, delivery status, retry policy, dan audit log.

## 9.6 Acceptance Criteria Analysis Cache

- Membuka analysis locked berulang kali tidak memanggil provider.
- Concurrent generation request menghasilkan satu logical generation.
- Cache hit dapat dibuktikan dengan provider call counter dan usage record.
- Failure tidak menghasilkan analysis locked palsu.
- Re-analysis membuat version baru tanpa merusak version lama.
- Hasil berstatus draft tidak dapat diekspor seolah SPK final.
- Admin tanpa permission tidak dapat memanggil generation endpoint.

---

# 10. Module UI/UX dan Navigation Blueprint

## 10.1 Global Navigation

Navigasi utama:

1. Overview
2. AI Triage Hub
3. Spatial & pHash Clustering
4. Deep Analysis & SPK
5. Lifecycle & Archive

Global UI elements:

- top bar: nama aplikasi, environment indicator, current admin, profile/logout;
- sidebar: lima module utama;
- global search report reference;
- notification/status center untuk worker, dead-letter, dan action pending;
- consistent filter/search/pagination pattern;
- visible loading, empty, error, permission-denied, dan stale-data state.

Jangan membuat seluruh informasi terlihat pada satu halaman panjang. Gunakan overview singkat dan detail per module.

## 10.2 Module 1 — Overview & System Health

### Tujuan

Memberi gambaran operasional, bukan menjadi tempat mengubah data utama.

### Layout yang Direkomendasikan

**Metric cards:**

- Total reports.
- Active reports.
- Pending triage.
- Reports in review.
- Processing reports.
- Completed waiting for archive.
- Archived reports.
- Duplicate submissions.

**System health:**

- API health.
- Database connectivity/latency percentile.
- Worker heartbeat status.
- Queue depth dan oldest queued age.
- AI provider status/circuit state.
- Daily/monthly estimated AI usage bila tersedia.

**Recent activity:**

- report submitted;
- triage succeeded/failed;
- duplicate relation confirmed;
- analysis generated/locked;
- lifecycle transition;
- report archived;
- admin action.

### Definisi Metric

- `Total Reports`: seluruh report record termasuk duplicate dan archived; jelaskan definisi ini di tooltip.
- `Active Reports`: report yang belum archived dan bukan duplicate.
- `Pending Triage`: triage status pending, bukan semata `report.status=submitted`.
- `Locked Analyses`: jumlah analysis version locked, bukan jumlah report yang dikunjungi.
- `Active Workers`: heartbeat worker yang masih berada dalam window valid.
- `DB Latency`: tampilkan p50/p95 atau hasil health query yang diberi label jelas; jangan hardcode angka demo seolah runtime metric.

### Data Loading

Gunakan aggregation query terindeks. Hindari menghitung semua event dan report row dari frontend.

Untuk awal, polling setiap 15–30 detik dapat digunakan. Ubah ke SSE ketika kebutuhan real-time terbukti. Aktivitas yang ditampilkan harus berasal dari audit/domain events, bukan parsing plain text log.

### Acceptance Criteria

- Metric mempunyai definisi konsisten.
- Nilai berubah setelah action nyata.
- Dashboard membedakan zero data dengan request failure.
- Worker offline dapat dibedakan dari queue kosong.
- Timestamp ditampilkan menurut timezone UI, tetapi disimpan UTC.

## 10.3 Module 2 — AI Triage & Worker Hub

### Tujuan

Memonitor dan mengendalikan background processing tanpa memberi akses AI langsung ke user.

### Layout

1. Worker health summary.
2. Queue summary: queued/running/retry/dead-letter.
3. Manual batch trigger.
4. Configured interval dan next schedule.
5. Queue table.
6. Batch execution history.
7. Per-report processing timeline.

### Queue Table Fields

- report reference;
- submitted at;
- lokasi ringkas;
- media count/preview;
- pHash short preview bila tersedia;
- triage status;
- attempt count;
- last error;
- action: view, retry, cancel bila state memungkinkan.

### Control Behavior

- `Trigger Manual Batch Worker` membuat batch/job, bukan menjalankan loop AI di HTTP handler.
- Button disabled saat batch identik sedang berjalan atau user tidak berpermission.
- Konfirmasi menampilkan jumlah kandidat dan estimasi policy cost jika tersedia.
- Interval setting memvalidasi batas minimum/maksimum.
- Perubahan interval disimpan sebagai configuration dengan audit trail.
- Manual retry hanya tersedia untuk job yang failed/dead-letter sesuai aturan.

### Worker Log View

Tampilkan event terstruktur: timestamp, batch ID, job ID, report ref, stage, duration, severity, message.

Contoh event stage:

`claimed` → `media_decoded` → `phash_calculated` → `provider_requested` → `output_validated` → `result_persisted` → `completed`.

Jangan menyimpan seluruh prompt atau image content ke log.

### Acceptance Criteria

- Admin dapat trigger batch dan melihat status asynchronous.
- API request cepat selesai dan tidak menunggu satu batch penuh.
- Job gagal dapat dilihat, di-retry, atau dead-letter.
- Retry tidak membuat triage run/vote ganda yang tidak disengaja.
- Worker heartbeat dan queue metrics berasal dari runtime.

## 10.4 Module 3 — Spatial & pHash Clustering

### Tujuan

Menjelaskan mengapa sebuah report menjadi duplicate/candidate dan memberi alat koreksi bagi admin.

### Layout

- map/list toggle;
- filter status, date range, category, confidence/match reason;
- cluster summary;
- master report detail;
- duplicate/candidate list;
- comparison panel dua report;
- map pin/cluster view;
- action confirm/reject/unlink.

### Map Behavior

- Map adalah visual aid, bukan sumber keputusan duplicate.
- Pin master dan duplicate dibedakan secara visual.
- Popup menampilkan distance dan relation status.
- Jangan menampilkan exact reporter identity kepada role yang tidak berhak.
- Saat map provider gagal, list view harus tetap dapat digunakan.
- Koordinat sensitif dapat dibulatkan pada role tertentu jika policy privasi membutuhkannya.

### Cluster Detail

Tampilkan:

- master report reference;
- title/summary;
- current status;
- upvote count dengan sumber yang jelas;
- related duplicates;
- spatial distance;
- pHash distance;
- algorithm version;
- decision reason;
- relation status dan audit history.

### Acceptance Criteria

- Cluster dapat dibuka dari map maupun list.
- Admin mengetahui alasan match.
- False positive dapat diputus.
- Map failure tidak memblokir operasi inti.
- Duplicate record masih dapat ditemukan melalui search/audit.

## 10.5 Module 4 — Deep Analysis & SPK Caching

### Tujuan

Menampilkan analisis teknis yang tersimpan, metadata cache, dan SPK draft dengan approval jelas.

### Layout

**Report selector:** reference, status, kategori, lokasi.  
**Report details:** image, description, location, upvotes, triage summary.  
**Analysis state:** not generated, pending, generating, locked, failed.  
**Analysis content:** technical assessment, restoration steps, risk notes, materials, labor, budget, assumptions.  
**SPK panel:** draft number, recipient team, scope, deadline proposal, approval status.  
**Actions:** generate, view version history, create new version, approve draft, export PDF.

### UX Rules

- Hasil locked ditampilkan sebagai cache hit.
- Tombol Generate tidak tampil aktif untuk role yang tidak berhak.
- Tampilkan warning ketika input report berubah setelah analysis dibuat.
- Tampilkan `Generated at`, `Prompt version`, `Model`, `Locked by`, `Version`.
- Hasil AI tidak diberi gaya visual seolah sudah diverifikasi engineer.
- Draft PDF diberi watermark dan metadata status.

### Acceptance Criteria

- Report selection hanya memicu read query.
- Generate adalah explicit action.
- Reopen tidak membuat API cost baru.
- Versi analysis dapat dilihat dan dibandingkan.
- SPK tidak dapat di-issue sebelum approval.

## 10.6 Module 5 — Lifecycle & Archive Manager

### Tujuan

Mengelola progress pekerjaan dan mempertahankan rekam jejak setelah report diarsipkan.

### Kanban Columns

- In Review.
- Processing.
- Completed — waiting H+1.

Archived berada di registry terpisah dan read-only.

### Action Behavior

Gunakan action button/menu per card dengan confirmation dan optional reason. Hindari drag-and-drop sebagai satu-satunya cara mengubah status karena lebih sulit menjamin validasi, permission, dan audit.

### Archive Registry

Fields:

- report reference;
- title/summary;
- completed at;
- archived at;
- final status;
- analysis/SPK version reference;
- action: view read-only, export history sesuai permission.

Archived record tidak menerima upvote/comment/status update dari jalur normal. Admin correction memerlukan explicit override yang tercatat.

### Acceptance Criteria

- Admin hanya dapat melakukan transisi valid.
- Completed report menampilkan countdown/eligible time berdasarkan completed_at.
- Archiver idempotent.
- Archived record read-only bagi API normal.
- Manual simulation tersedia hanya untuk development/demo atau permission khusus.

---

# 11. API Contract Blueprint

Gunakan prefix versioning `/api/v1`. Seluruh endpoint admin memerlukan authentication dan authorization. Nama berikut merupakan rancangan kontrak; sesuaikan dengan konvensi codebase yang sudah ada, tetapi jangan mengubah semantik bisnisnya.

## 11.1 Public/User API

| Method | Endpoint | Tujuan |
|---|---|---|
| POST | `/api/v1/reports` | Submit report tanpa AI call |
| GET | `/api/v1/reports/{id}` | Membaca status report milik pemohon sesuai access policy |
| POST | `/api/v1/uploads/init` | Membuat upload authorization jika direct upload dipakai |
| POST | `/api/v1/uploads/{id}/finalize` | Memverifikasi upload sebelum report dibuat |

Public API tidak menyediakan endpoint `generate-triage` atau `generate-analysis`.

## 11.2 Admin Overview

| Method | Endpoint | Tujuan |
|---|---|---|
| GET | `/api/v1/admin/dashboard/summary` | Aggregated metrics |
| GET | `/api/v1/admin/dashboard/activity` | Recent audit/domain activity |
| GET | `/api/v1/admin/system/health` | API, database, worker, queue health |

## 11.3 Admin Triage

| Method | Endpoint | Tujuan |
|---|---|---|
| GET | `/api/v1/admin/triage/queue` | Queue listing dan pagination |
| POST | `/api/v1/admin/triage/batches` | Trigger manual batch |
| GET | `/api/v1/admin/triage/batches/{id}` | Batch progress |
| POST | `/api/v1/admin/reports/{id}/triage/retry` | Retry report triage |
| GET | `/api/v1/admin/jobs` | Filter job state |
| GET | `/api/v1/admin/jobs/{id}` | Job detail/log timeline |

## 11.4 Admin Clustering

| Method | Endpoint | Tujuan |
|---|---|---|
| GET | `/api/v1/admin/clusters` | Cluster listing |
| GET | `/api/v1/admin/reports/{id}/relations` | Report relationship list |
| POST | `/api/v1/admin/relations/{id}/confirm` | Confirm possible duplicate |
| POST | `/api/v1/admin/relations/{id}/reject` | Reject candidate |
| POST | `/api/v1/admin/relations/{id}/reverse` | Reverse confirmed relation dengan alasan |

## 11.5 Admin Deep Analysis dan SPK

| Method | Endpoint | Tujuan |
|---|---|---|
| GET | `/api/v1/admin/reports/{id}/analyses` | List analysis versions/cache state |
| POST | `/api/v1/admin/reports/{id}/analyses` | Request generation versi baru |
| GET | `/api/v1/admin/analyses/{id}` | Read immutable analysis |
| POST | `/api/v1/admin/analyses/{id}/approve` | Approve analysis/SPK sesuai permission |
| POST | `/api/v1/admin/analyses/{id}/exports/pdf` | Create PDF export |

## 11.6 Lifecycle dan Archive

| Method | Endpoint | Tujuan |
|---|---|---|
| GET | `/api/v1/admin/reports` | Report listing dengan filter |
| POST | `/api/v1/admin/reports/{id}/transitions` | Request state transition |
| GET | `/api/v1/admin/archive` | Archived registry |
| GET | `/api/v1/admin/archive/{id}` | Read-only archived detail |
| POST | `/api/v1/admin/archive/simulations` | Simulasi archiver pada demo/dev environment saja |

## 11.7 API Contract Rules

- Pagination menggunakan cursor atau pagination konsisten; jangan memuat seluruh laporan sekaligus.
- Filter status/category/date harus tervalidasi.
- Gunakan request ID/correlation ID.
- Gunakan idempotency key untuk submission, batch trigger, analysis generation, dan action yang dapat diulang.
- Error response mempunyai machine-readable code, safe message, dan request ID.
- Jangan mengembalikan stack trace, SQL detail, prompt internal, atau secret.
- Semua timestamp dikirim sebagai ISO-8601 dengan timezone.
- Gunakan optimistic concurrency/version field atau expected status saat transition agar stale UI tidak menimpa update terbaru.

---

# 12. Authentication, Authorization, dan Security

## 12.1 Authentication

Pilih satu mekanisme yang sesuai deployment: secure server session atau token-based authentication. Jangan mencampur beberapa pola tanpa alasan.

Untuk browser admin portal, cookie session `HttpOnly`, `Secure`, dan `SameSite` biasanya mengurangi eksposur token ke JavaScript. Jika cookie digunakan, proteksi CSRF wajib dipertimbangkan.

## 12.2 Role-Based Access Control

Contoh permission matrix awal:

| Action | Admin | Facility Manager | Technician | Reviewer |
|---|---:|---:|---:|---:|
| View operational dashboard | Ya | Ya | Terbatas | Ya |
| View report details | Ya | Ya | Assigned | Ya |
| Trigger triage batch | Ya | Opsional | Tidak | Tidak |
| Retry failed triage | Ya | Opsional | Tidak | Tidak |
| Generate deep analysis | Ya | Sesuai policy | Tidak | Tidak |
| Approve SPK | Ya/authorized approver | Ya bila ditunjuk | Tidak | Opsional |
| Transition to processing | Ya | Ya | Sesuai assignment | Tidak |
| Mark completed | Ya | Ya | Sesuai assignment | Tidak |
| Reverse duplicate relation | Ya | Opsional | Tidak | Tidak |
| Archive override | Sangat terbatas | Tidak | Tidak | Tidak |

V1 boleh hanya mengaktifkan `admin` dan `technician` bila organisasi masih kecil. Namun, permission check tetap dibuat per action, bukan hanya menyembunyikan tombol di frontend.

## 12.3 Security Controls

- Password hashing menggunakan algoritma password-hashing yang sesuai, bukan enkripsi reversible.
- TLS/HTTPS untuk seluruh trafik.
- Secret disimpan di environment/secret manager, bukan repository.
- Rate limit untuk login, public submission, upload, dan admin generation.
- Request body dan upload size limit.
- Input validation dan output encoding.
- CORS allowlist; jangan gunakan wildcard untuk credentialed admin request.
- CSRF protection bila memakai cookie session.
- SQL query memakai parameterized query.
- Audit untuk admin action sensitif.
- Jangan expose object storage bucket secara public.
- Signed URL mempunyai masa berlaku singkat.
- Dependency update dan vulnerability scan dalam CI.
- Backup database dan uji restore.

## 12.4 Privacy

- Ambil data personal minimum yang dibutuhkan.
- Batasi visibilitas reporter identity berdasarkan role.
- Tentukan retention policy untuk original image, thumbnail, AI output, dan audit.
- Jangan mengirim data personal yang tidak relevan ke AI provider.
- Jangan log full GPS, private media URL, token, atau raw prompt sensitif tanpa kebutuhan.
- Catat consent/notice bila kebijakan organisasi mensyaratkannya.

---

# 13. Lifecycle dan H+1 Auto-Archiver

## 13.1 Definisi H+1

Untuk V1, H+1 berarti **24 jam penuh setelah timestamp laporan berubah menjadi completed**, bukan pergantian tanggal kalender.

Cutoff dihitung dari `completed_at + 24 hours` dalam waktu UTC. UI boleh menampilkan waktu lokal Asia/Jakarta.

## 13.2 Kenapa Bukan `updated_at`

`updated_at` berubah ketika metadata, note, assignment, atau informasi lain diperbarui. Jika archiver memakai `updated_at`, laporan dapat tertunda diarsipkan tanpa sengaja atau justru diarsipkan berdasarkan perubahan yang tidak terkait penyelesaian.

Gunakan `completed_at` sebagai sumber waktu lifecycle.

## 13.3 Archiver Workflow

1. Scheduler membuat archive sweep job berdasarkan interval konfigurasi.
2. Worker mencari report berstatus completed dengan `completed_at` sudah melewati cutoff.
3. Worker claim/lock kandidat agar dua worker tidak memproses record yang sama secara tidak aman.
4. Worker melakukan conditional transition: hanya update bila status masih completed dan cutoff masih terpenuhi.
5. Worker mengisi `archived_at`.
6. Worker menulis audit/domain event.
7. Job menyimpan count success/skipped/failed.
8. Pengulangan sweep tidak mengubah hasil atau menggandakan audit secara tidak terkendali.

## 13.4 Scheduler Reliability

Ticker dalam process dapat cukup untuk satu instance, tetapi beberapa replica dapat menjalankan scheduler bersamaan. Karena itu:

- gunakan database lease/advisory lock untuk satu logical schedule;
- atau jalankan scheduler hanya pada satu worker deployment;
- tetap buat archiver idempotent;
- jangan mengandalkan process-local timer sebagai satu-satunya jaminan job berjalan.

## 13.5 Archive Freeze

Setelah archived:

- public upvote/comment ditolak;
- normal status transition ditolak;
- analisis dan SPK version yang ada tetap dapat dibaca sesuai permission;
- export history boleh dilakukan;
- correction/reopen hanya lewat administrative override dengan reason dan audit.

## 13.6 Simulation Mode

`Run H+1 Cron Job Simulation` harus jelas sebagai mode demo.

- Hanya tersedia di development/demo atau admin permission khusus.
- Tidak boleh mempercepat waktu data production tanpa audit.
- Lebih aman menggunakan seeded report dengan `completed_at` yang sengaja diatur lebih dari 24 jam.
- UI menampilkan label `Simulation` agar evaluator tidak mengira scheduler production benar-benar dipaksa berjalan.

---

# 14. Observability dan Operational Readiness

## 14.1 Structured Logging

Setiap log minimal memuat:

- timestamp;
- severity;
- service/process name;
- environment;
- request ID atau job ID;
- report ID bila relevan;
- operation/stage;
- duration;
- error code.

Jangan masukkan password, token, API key, signed URL, image content, atau data personal yang tidak dibutuhkan.

## 14.2 Metrics Minimum

### API

- request count;
- error rate;
- latency p50/p95/p99;
- request size;
- rate-limit hits.

### Database

- connection pool usage;
- query latency;
- slow query count;
- transaction failure/deadlock;
- database connectivity.

### Worker

- queue depth;
- oldest queued age;
- job success/failure/retry;
- processing duration;
- lease expired count;
- dead-letter count;
- worker heartbeat age.

### AI

- calls by task/provider/model;
- success/failure/unknown outcome;
- latency;
- usage/cost;
- cache hit rate;
- provider rate-limit/error count.

### Domain

- reports submitted;
- duplicate rate;
- triage completion rate;
- manual correction rate;
- average time to review;
- average time to completion;
- archive lag.

## 14.3 Health Check Levels

- Liveness: process hidup.
- Readiness: process siap melayani request, dependency wajib tersedia.
- Worker heartbeat: worker loop aktif.
- Provider readiness: status provider atau circuit state; provider outage tidak harus membuat API intake menjadi unavailable.

Pisahkan API readiness dari AI provider readiness. Ingestion harus tetap tersedia meskipun AI vendor sedang down.

## 14.4 Alerts Awal

- worker heartbeat stale;
- oldest queue age melewati ambang;
- dead-letter job bertambah;
- AI budget threshold tercapai;
- provider error rate tinggi;
- database connection pool jenuh;
- archive lag melewati toleransi;
- repeated auth failures.

Ambang alert dikalibrasi berdasarkan traffic dan target organisasi, bukan angka asumsi yang dianggap universal.

---

# 15. Testing Strategy

Testing harus membuktikan business invariant, bukan hanya menambah jumlah test.

## 15.1 Unit Tests

- report transition matrix;
- permission matrix;
- severity/schema validation;
- pHash parser dan Hamming Distance;
- spatial threshold policy (menggunakan nilai hasil repository mock);
- cache key/input fingerprint;
- retry/backoff policy;
- idempotency key generation;
- archive cutoff calculation;
- budget guard.

## 15.2 Integration Tests

Gunakan PostgreSQL dengan PostGIS yang nyata untuk menguji:

- `ST_DWithin` dan geography meter semantics;
- GiST index/query plan bila relevan;
- transaction rollback;
- unique idempotency constraints;
- duplicate concurrent insertion;
- job claim dengan beberapa worker;
- lease expiry dan reclaim;
- archive conditional update;
- foreign key/cascade policy.

Jangan mengganti PostGIS dengan mock untuk test yang ingin membuktikan geospatial behavior.

## 15.3 AI Provider Tests

Gunakan fake provider yang dapat:

- mengembalikan response valid;
- mengembalikan malformed JSON;
- timeout;
- rate limit;
- provider 5xx;
- menghasilkan output di luar schema;
- menambahkan usage metadata;
- menghitung jumlah invocation.

Test wajib membuktikan ingestion tidak memanggil AI, retry sesuai policy, dan cached analysis tidak membuat invocation baru.

## 15.4 Concurrency Tests

- dua submission serupa masuk bersamaan;
- dua worker claim job bersamaan;
- dua admin meminta deep analysis bersamaan;
- dua admin melakukan status transition bersamaan;
- duplicate auto-vote diproses berulang;
- archive sweep dijalankan oleh lebih dari satu worker.

Gunakan race detector untuk Go code yang relevan, tetapi pahami bahwa race detector tidak menggantikan database concurrency tests.

## 15.5 API/Frontend E2E Tests

Happy path:

1. User submit report.
2. Admin melihat queue.
3. Worker triage.
4. Admin melihat hasil triage.
5. Admin meminta deep analysis.
6. Worker menyimpan analysis locked.
7. Admin membuka kembali report dan melihat cache hit.
8. Admin mengubah lifecycle.
9. Archiver memindahkan report setelah eligible.

Failure path:

- AI unavailable;
- invalid image;
- duplicate false positive;
- duplicate worker retry;
- unauthorized deep analysis;
- concurrent generation;
- analysis malformed;
- archive job berjalan dua kali.

## 15.6 Load dan Resilience Test

Mulai dengan beban realistis, bukan benchmark besar tanpa profil penggunaan.

Uji:

- burst submission tanpa AI;
- queue backlog dan recovery;
- image decode resource limit;
- provider rate-limit;
- database unavailable sementara;
- worker restart saat job running;
- object storage timeout;
- recovery setelah deployment/restart.

---

# 16. Implementation Roadmap — Urutan Pengerjaan Detail

Urutan ini disusun berdasarkan dependency. Jangan mulai dari animasi dashboard atau map styling sebelum contract dan data state stabil.

## Phase 0 — Requirement Freeze dan Architecture Decision

**Tujuan:** mengunci aturan yang jika berubah di tengah jalan akan menyebabkan rewrite besar.

### Tasks

1. Tetapkan definisi report status, triage status, dan analysis status.
2. Tetapkan apakah duplicate submission disimpan. Rekomendasi blueprint: simpan sebagai record dan relasikan ke master.
3. Tetapkan default duplicate rule: spatial radius 15m dan pHash Hamming Distance `<= 5` sebagai initial configuration; kedua sinyal diperlukan untuk auto-link.
4. Tetapkan definisi H+1: 24 jam setelah `completed_at`.
5. Tetapkan role minimum dan permission untuk generate analysis, approve SPK, transition, dan archive override.
6. Tetapkan format media dan limit ukuran awal.
7. Tetapkan provider AI awal, model, timeout, retry, dan budget guard.
8. Tetapkan apakah tenant scope dibutuhkan V1; siapkan nullable tenant key bila masuk akal.
9. Buat architecture decision record (ADR) untuk keputusan yang berdampak besar.
10. Definisikan API error format, pagination, timestamp format, dan idempotency header.

### Deliverables

- state transition table;
- ERD awal;
- endpoint contract;
- configuration list;
- ADR untuk duplicate policy, queue, analysis lock, archive time.

### Exit Criteria

Tidak ada ambiguitas tentang status, duplicate persistence, cache semantics, atau H+1.

## Phase 1 — Repository Foundation dan Runtime Configuration

**Tujuan:** menyiapkan struktur codebase tanpa menambah kompleksitas deployment.

### Tasks

1. Pisahkan package domain, application/use case, repository, transport/HTTP, worker, dan infrastructure adapter.
2. Buat process mode API dan worker dari codebase yang sama.
3. Buat configuration loader dengan validation saat startup.
4. Pisahkan environment: local, test, demo/staging, production.
5. Tambahkan graceful shutdown untuk HTTP server dan worker.
6. Tambahkan request ID/correlation ID middleware.
7. Tambahkan structured logger.
8. Tambahkan liveness/readiness endpoint.
9. Tambahkan migration runner dan migration versioning.
10. Tambahkan CI baseline: formatting, lint, unit test, build.

### Deliverables

- app dapat start dalam API mode;
- app dapat start dalam worker mode;
- config invalid gagal saat startup dengan error jelas;
- migration dapat dijalankan secara repeatable.

### Exit Criteria

API dan worker dapat dijalankan terpisah, shutdown tidak meninggalkan goroutine/job lease tanpa penanganan.

## Phase 2 — Database Schema dan Repository Layer

**Tujuan:** menyelesaikan data contract sebelum business workflow dibangun.

### Tasks

1. Buat migration users/roles.
2. Buat migration reports.
3. Aktifkan PostGIS extension melalui deployment/migration policy.
4. Buat report_media.
5. Buat report_relations dan report_votes.
6. Buat triage_runs.
7. Buat analyses/analysis_contents.
8. Buat jobs dan worker_heartbeats.
9. Buat audit_logs.
10. Tambahkan FK, unique, check, dan spatial indexes.
11. Implementasikan repository method yang sesuai query use case.
12. Tambahkan repository integration tests.
13. Jalankan query plan terhadap spatial search dan job claim query.

### Deliverables

- schema V1;
- migration up/down yang aman sesuai kebijakan;
- repository tests;
- ERD yang diperbarui.

### Exit Criteria

Database constraints menolak data invalid dan repository dapat menjalankan query inti tanpa full table scan yang tidak perlu.

## Phase 3 — Report Intake dan Media Storage

**Tujuan:** membuat jalur submission stabil dan bebas AI.

### Tasks

1. Implementasikan request DTO dan validation.
2. Implementasikan report creation service.
3. Implementasikan idempotency key untuk submission.
4. Integrasikan media storage adapter.
5. Implementasikan MIME sniffing, file size/dimension limit, dan checksum.
6. Simpan report + media metadata secara konsisten.
7. Masukkan job preprocessing/dedupe setelah commit melalui transactional job insert atau outbox pattern.
8. Buat orphan object cleanup strategy.
9. Implementasikan rate limit untuk public intake.
10. Buat endpoint status report.
11. Tambahkan fake AI provider call counter untuk membuktikan intake tidak memanggil AI.

### Deliverables

- public submission API;
- media object storage integration;
- idempotent submission;
- tests untuk storage failure dan invalid file.

### Exit Criteria

Submission tetap berhasil saat AI provider down dan tidak ada AI invocation pada intake path.

## Phase 4 — Dedupe Engine (PostGIS + pHash)

**Tujuan:** membangun duplicate detection yang terukur dan dapat dikoreksi.

### Tasks

1. Pilih library pHash yang dapat dipakai dan lisensinya sesuai.
2. Tetapkan format hash dan algorithm version.
3. Implementasikan image decode dengan batas resource.
4. Implementasikan pHash extraction.
5. Implementasikan Hamming Distance.
6. Implementasikan spatial candidate query dengan PostGIS geography.
7. Jadikan radius dan threshold configurable.
8. Implementasikan match decision policy.
9. Implementasikan relation creation dan duplicate status.
10. Implementasikan idempotent auto-vote.
11. Implementasikan candidate review state.
12. Tambahkan spatial tile/advisory lock atau concurrency guard.
13. Buat admin endpoint untuk confirm/reject/reverse.
14. Buat tests untuk boundary radius, threshold, incompatible hash version, dan concurrent submission.

### Deliverables

- dedupe service;
- relation history;
- cluster API;
- explanation metadata.

### Exit Criteria

Duplicate dapat dideteksi, alasan dapat ditampilkan, false positive dapat dibatalkan, dan worker retry tidak menggandakan vote.

## Phase 5 — Durable Job Queue dan Worker Runtime

**Tujuan:** menjalankan proses asynchronous tanpa kehilangan job saat restart.

### Tasks

1. Implementasikan atomic job claim.
2. Implementasikan lease dan lease renewal.
3. Implementasikan stale lease recovery.
4. Implementasikan retry/backoff/jitter.
5. Implementasikan dead-letter state.
6. Implementasikan job idempotency.
7. Implementasikan batch grouping.
8. Implementasikan worker heartbeat.
9. Implementasikan queue metrics.
10. Implementasikan manual retry/cancel bila valid.
11. Pisahkan concurrency limit berdasarkan job type.
12. Tambahkan graceful shutdown dan lease release policy.
13. Uji worker restart ketika job sedang berjalan.

### Deliverables

- worker runtime;
- jobs API;
- heartbeat;
- retry/dead-letter view data.

### Exit Criteria

Job tidak hilang saat process restart, tidak diproses dua kali tanpa proteksi idempotency, dan kegagalan dapat ditangani operator.

## Phase 6 — AI Triage Integration

**Tujuan:** memproses report canonical secara asynchronous dan menyimpan hasil terstruktur.

### Tasks

1. Definisikan triage provider interface.
2. Implementasikan adapter provider pertama.
3. Definisikan prompt ID dan prompt version.
4. Definisikan structured output schema.
5. Implementasikan input fingerprint.
6. Implementasikan provider timeout dan error classification.
7. Implementasikan usage/cost capture.
8. Implementasikan triage job handler.
9. Validasi response sebelum disimpan.
10. Simpan triage run dan ringkasan report.
11. Transisikan report ke in_review hanya setelah kondisi yang sesuai.
12. Implementasikan manual retry dan failed state.
13. Tambahkan budget/rate guard.
14. Buat fake provider tests.

### Deliverables

- AI triage adapter;
- triage worker handler;
- triage history;
- structured output validation.

### Exit Criteria

Triage tidak berjalan dalam HTTP request user, output invalid tidak masuk ke report sebagai hasil sukses, dan AI outage tidak menghentikan intake.

## Phase 7 — Admin API dan Overview

**Tujuan:** membuat data operasional dapat dibaca admin sebelum seluruh UI detail selesai.

### Tasks

1. Implementasikan admin authentication.
2. Implementasikan role/permission middleware.
3. Implementasikan dashboard aggregation query.
4. Implementasikan activity feed dari audit/domain event.
5. Implementasikan health endpoint untuk database/worker/queue/provider.
6. Implementasikan report list/filter/pagination.
7. Implementasikan report detail query.
8. Tambahkan caching read-only hanya bila query telah terbukti mahal.
9. Uji metric definition terhadap fixture dataset.

### Deliverables

- dashboard summary API;
- activity API;
- admin report API;
- system health API.

### Exit Criteria

Dashboard menampilkan angka yang dapat direkonsiliasi dengan database dan tidak mengarang status online.

## Phase 8 — Admin UI: Overview dan Triage Hub

**Tujuan:** menyelesaikan dua module yang memperlihatkan operasi utama.

### Tasks

1. Buat app shell, sidebar, topbar, route guard.
2. Buat reusable metric card, table, badge, filter, pagination.
3. Buat Overview page.
4. Buat worker status panel.
5. Buat queue list dan batch history.
6. Buat manual batch trigger dengan confirmation.
7. Buat job detail/timeline.
8. Buat loading/empty/error/stale state.
9. Tambahkan polling terukur dan cancellation saat page unmount.
10. Pastikan UI tidak menampilkan API secret atau internal payload.

### Deliverables

- Overview usable;
- Triage Hub usable;
- admin dapat mengoperasikan batch dan memeriksa hasil.

### Exit Criteria

Operator dapat menemukan pending, failed, retry, dan completed triage tanpa membuka database secara manual.

## Phase 9 — Spatial & pHash Clustering UI

**Tujuan:** membuat duplicate decision transparan.

### Tasks

1. Implementasikan cluster/list API.
2. Pilih map component/provider.
3. Buat map/list toggle.
4. Buat cluster detail panel.
5. Buat report comparison panel.
6. Tampilkan spatial distance dan pHash distance.
7. Implementasikan confirm/reject/reverse action.
8. Tampilkan relation audit history.
9. Tambahkan list fallback jika map provider gagal.
10. Uji responsive layout dan permission.

### Deliverables

- cluster map/list;
- duplicate review actions;
- relation explanation.

### Exit Criteria

Admin dapat memahami dan mengoreksi keputusan duplicate tanpa bergantung pada log backend.

## Phase 10 — Deep Analysis, Cache Guard, dan SPK

**Tujuan:** menambahkan demonstrasi high-tier AI yang cost-aware.

### Tasks

1. Definisikan deep analysis schema.
2. Definisikan provider adapter dan prompt version terpisah dari triage.
3. Implementasikan analysis request service.
4. Implementasikan fingerprint dan cache lookup.
5. Implementasikan active-generation uniqueness/idempotency.
6. Implementasikan deep analysis job handler.
7. Implementasikan unknown provider outcome handling.
8. Validasi technical steps, estimates, dan SPK fields.
9. Simpan analysis content sebagai immutable version.
10. Implementasikan version history.
11. Implementasikan approval workflow minimum.
12. Implementasikan PDF draft/export.
13. Tambahkan fake provider tests untuk cache hit dan concurrent request.

### Deliverables

- analysis generation API;
- immutable analysis version;
- cache hit indicator;
- SPK draft/export.

### Exit Criteria

Dua atau lebih request/view atas analysis locked tidak menghasilkan provider call tambahan. Re-analysis membuat versi baru dan tercatat.

## Phase 11 — Lifecycle, Archive, dan H+1

**Tujuan:** mengelola status kerja dan archive policy secara konsisten.

### Tasks

1. Implementasikan lifecycle service dan transition matrix.
2. Implementasikan expected-status/version check.
3. Implementasikan audit log untuk transition.
4. Simpan `completed_at` hanya pada transisi ke completed.
5. Implementasikan scheduler/archive job.
6. Implementasikan worker lock/claim untuk kandidat archive.
7. Implementasikan conditional update terhadap completed_at cutoff.
8. Implementasikan archived read-only API policy.
9. Implementasikan archive registry UI.
10. Implementasikan simulation mode hanya untuk demo/dev.
11. Uji dua scheduler/worker concurrent.

### Deliverables

- lifecycle API;
- Kanban/board;
- archive registry;
- H+1 worker.

### Exit Criteria

Report tidak diarsipkan sebelum 24 jam dari completed_at dan archive job aman ketika dijalankan berulang.

## Phase 12 — Security, Resilience, dan Operational Hardening

**Tujuan:** menutup celah fatal sebelum demo dinilai sebagai sistem yang dapat digunakan.

### Tasks

1. Review authentication dan session/token handling.
2. Review authorization setiap endpoint admin.
3. Tambahkan CSRF protection bila diperlukan.
4. Review CORS dan security headers.
5. Review upload validation dan object access.
6. Review rate limiting dan abuse path.
7. Review secret handling.
8. Review prompt injection/data exposure.
9. Review log redaction.
10. Tambahkan backup dan restore procedure.
11. Tambahkan DB migration rollback/recovery notes.
12. Uji worker/provider/storage/database failure.
13. Tambahkan alert minimum.
14. Jalankan dependency/security scan.
15. Jalankan end-to-end dan concurrency suite.

### Deliverables

- security checklist;
- operational runbook;
- backup/restore evidence;
- failure test report.

### Exit Criteria

Tidak ada endpoint AI publik, secret tidak bocor, unauthorized admin action ditolak, dan sistem dapat pulih dari kegagalan transient.

## Phase 13 — Demo Dataset, Documentation, dan Release V1

**Tujuan:** membuat demo akademik dapat diulang dan tidak bergantung pada data acak.

### Tasks

1. Buat seed dataset dengan report untuk semua status.
2. Buat dua report duplicate dengan jarak dan hash yang diketahui.
3. Buat satu possible duplicate yang hanya cocok pada satu sinyal.
4. Buat worker log/fixture yang merepresentasikan event nyata atau secara eksplisit bertanda simulation.
5. Buat satu report dengan triage success.
6. Buat satu report dengan triage failure untuk menunjukkan retry/dead-letter.
7. Buat analysis locked dengan usage metadata.
8. Buat archived report dengan completed_at lebih dari 24 jam.
9. Buat script/demo procedure yang deterministic.
10. Siapkan diagram arsitektur dan ERD.
11. Siapkan API documentation.
12. Siapkan README setup dan runbook.
13. Jalankan full test suite dari environment bersih.
14. Buat release tag dan changelog.

### Deliverables

- V1 demo environment;
- reproducible seed data;
- demo script;
- architecture pack;
- release notes.

### Exit Criteria

Evaluator dapat mengikuti alur utama secara konsisten dan dapat membedakan data runtime, cached AI result, dan simulation fixture.

---

# 17. Prioritas Release: Must Have vs Later

## 17.1 V1 Must Have

- Report intake tanpa AI.
- Media storage dan validasi dasar.
- PostGIS spatial candidate search.
- pHash dan Hamming Distance.
- Duplicate relation yang tidak menghapus submission.
- Durable database-backed job queue.
- Worker retry/lease/idempotency.
- AI triage asynchronous.
- Admin authentication dan permission minimum.
- Overview metric yang benar.
- Triage Hub dengan manual trigger dan error visibility.
- Deep analysis explicit request + immutable cache version.
- Cache hit tanpa provider call.
- Lifecycle transition yang tervalidasi.
- H+1 archive berbasis completed_at.
- Audit log untuk admin actions.
- Tests untuk race condition utama.

## 17.2 V1 Should Have

- Cluster map dengan list fallback.
- SPK draft PDF export.
- Analysis version comparison.
- Cost/usage dashboard.
- SSE activity update.
- Tenant scope jika memang dibutuhkan.
- Automated object cleanup.

## 17.3 Later

- WhatsApp official messaging integration.
- External message delivery tracking.
- Separate message broker.
- Separate AI service.
- Advanced semantic duplicate model.
- Fine-tuned vision model.
- SLA prediction dan predictive maintenance.
- Procurement integration.
- Multi-organization billing/quota.

---

# 18. Demo Blueprint untuk Academic Review

Demo harus memperlihatkan sebab-akibat arsitektur, bukan hanya klik antarhalaman.

## Demo 1 — Zero-AI-Cost Ingestion

**Setup:** fake provider memiliki invocation counter.

1. User mengirim laporan baru beserta gambar dan koordinat.
2. API mengembalikan report reference dengan status submitted.
3. Admin membuka Triage Hub.
4. Report terlihat pending.
5. Tunjukkan provider invocation counter tetap nol sebelum worker menjalankan triage.

**Claim yang dapat dipertanggungjawabkan:** jalur ingestion menyimpan data dan membuat job tanpa memanggil AI.

## Demo 2 — Duplicate Detection

**Setup:** seeded master report dan dua submission terkait.

1. Buka Spatial & pHash Clustering.
2. Pilih master report.
3. Tampilkan duplicate dengan jarak lokasi dan Hamming Distance.
4. Tampilkan auto-upvote event yang berasal dari duplicate submission.
5. Tampilkan possible duplicate dengan hanya satu sinyal cocok.
6. Tolak/konfirmasi candidate dan tunjukkan audit.

**Claim yang dapat dipertanggungjawabkan:** PostGIS mempersempit kandidat geografis; pHash membandingkan kemiripan visual; sistem tidak mengandalkan AI untuk pemeriksaan similarity dasar.

## Demo 3 — Async AI Triage

1. Trigger manual batch.
2. UI menerima batch ID tanpa menunggu semua job selesai.
3. Tunjukkan job lifecycle dan worker logs.
4. Tunjukkan triage summary, category, severity, model, dan prompt version setelah berhasil.
5. Simulasikan provider failure menggunakan fake provider.
6. Tunjukkan retry/dead-letter state.

**Claim yang dapat dipertanggungjawabkan:** AI triage dijalankan asynchronous dan kegagalan provider tidak membatalkan submission.

## Demo 4 — Deep Analysis Cache

1. Pilih report yang belum dianalisis.
2. Tampilkan state `Not Generated`.
3. Admin meminta generation.
4. Tampilkan generation job dan hasil locked.
5. Catat provider invocation count.
6. Refresh halaman atau buka report kembali beberapa kali.
7. Tunjukkan cache hit dan counter tidak bertambah.
8. Tampilkan version/prompt/model metadata.

**Claim yang dapat dipertanggungjawabkan:** hasil deep analysis tersimpan sebagai immutable snapshot dan tidak memanggil provider ketika dibaca kembali.

## Demo 5 — Lifecycle dan H+1 Archive

1. Ubah report dari in_review ke processing.
2. Selesaikan report melalui authorized transition.
3. Tampilkan `completed_at`.
4. Jalankan archive simulation pada seeded record yang sudah melewati 24 jam.
5. Tunjukkan status archived dan interaksi normal ditolak.
6. Tunjukkan audit trail.

**Claim yang dapat dipertanggungjawabkan:** lifecycle dikendalikan state machine dan archive job berjalan berdasarkan timestamp penyelesaian.

## 18.1 Demo Honesty Rules

- Jangan tampilkan log fixture sebagai live log.
- Jangan menyebut angka metric sebagai real-time jika berasal dari seed.
- Jangan menyebut cache hit sebagai zero-cost total system; sebut zero additional AI provider call.
- Jangan mengklaim AI selalu benar; tunjukkan human review dan correction path.
- Jangan mengklaim PostGIS/pHash mengetahui kerusakan yang sama secara semantik; keduanya memberi sinyal lokasi dan visual.
- Jangan menyebut SPK AI sebagai dokumen final sebelum approval.

---

# 19. Risiko Fatal dan Mitigasi

| Risiko | Dampak | Mitigasi wajib |
|---|---|---|
| AI dipanggil dari intake handler | Latency, biaya, ingestion gagal saat provider down | Pisahkan intake dari worker dan gunakan fake provider test |
| Queue hanya in-memory | Job hilang saat restart | Database-backed queue/outbox |
| Retry tanpa idempotency | Duplicate triage, upvote, atau biaya AI | Unique key, fingerprint, handler idempotency |
| Dua worker membuat duplicate master | Cluster tidak konsisten | Transaction lock/claim dan concurrency test |
| Auto-merge berdasarkan satu sinyal | False positive menggabungkan masalah berbeda | Require spatial + pHash match untuk auto-link awal |
| Duplicate submission dihapus | Bukti dan audit pelapor hilang | Simpan report record dan relation |
| `updated_at` dipakai untuk H+1 | Archive timing berubah karena update non-lifecycle | Gunakan `completed_at` |
| Concurrent deep analysis | Double API cost | Active-generation uniqueness dan idempotency |
| Cache guard hanya di frontend | Dapat dilewati dan tetap menimbulkan biaya | Cache guard di backend/database |
| Mengedit locked analysis | Riwayat dan audit hilang | Immutable versioning |
| AI output dianggap instruksi final | Risiko teknis dan keselamatan | Human approval, uncertainty, safety review |
| Provider timeout langsung retry | Double charge jika provider sudah memproses | Unknown outcome/reconciliation policy |
| Worker ONLINE hanya dari API health | False health status | Worker heartbeat dan queue metrics |
| Dashboard memakai angka hardcoded | Demo menyesatkan | Runtime query atau label simulation |
| Map menjadi dependency wajib | Admin tidak dapat bekerja saat map down | List fallback |
| Status diubah langsung dari frontend | Invalid transition/race | Lifecycle service dan permission check |
| File media public permanen | Kebocoran bukti/lokasi | Private bucket dan signed URL |
| Tanpa backup restore test | Data tidak dapat dipulihkan | Backup schedule dan restore drill |

---

# 20. Configuration Blueprint

Seluruh nilai yang dapat berubah karena deployment atau kalibrasi harus berada di configuration, bukan hardcoded di business logic.

| Configuration | Contoh nilai awal | Catatan |
|---|---:|---|
| `DEDUPE_RADIUS_METERS` | 15 | Kalibrasi dengan data lapangan |
| `PHASH_MAX_DISTANCE` | 5 | Hanya untuk hash version kompatibel |
| `PHASH_ALGORITHM_VERSION` | v1 | Simpan pada setiap media hash |
| `TRIAGE_BATCH_SIZE` | 5–20 | Sesuaikan provider quota dan latency |
| `TRIAGE_MAX_ATTEMPTS` | 3 | Tidak untuk semua error |
| `TRIAGE_TIMEOUT` | Configurable | Berdasarkan provider/model |
| `WORKER_CONCURRENCY_TRIAGE` | 1–N | Mulai konservatif |
| `WORKER_CONCURRENCY_ANALYSIS` | 1 | Jaga cost dan concurrent generation |
| `JOB_LEASE_DURATION` | Configurable | Lebih panjang dari operasi normal dengan renewal |
| `JOB_RETRY_BASE_DELAY` | Configurable | Exponential backoff + jitter |
| `H1_ARCHIVE_HOURS` | 24 | Definisi H+1 V1 |
| `DASHBOARD_POLL_INTERVAL` | 15–30 seconds | UI refresh, bukan scheduler worker |
| `MAX_UPLOAD_BYTES` | Ditentukan produk | Harus konsisten dengan infra |
| `SIGNED_URL_TTL` | Short-lived | Jangan terlalu panjang |
| `AI_DAILY_BUDGET` | Ditentukan pemilik sistem | Fail-safe saat budget habis |

Nilai di atas adalah titik awal untuk konfigurasi, bukan jaminan performa atau rekomendasi universal.

---

# 21. Deployment dan Environment

## 21.1 Local Development

Minimum dependency:

- Go API process;
- Go worker process;
- PostgreSQL + PostGIS;
- local object storage emulator atau filesystem adapter khusus local;
- fake AI provider default;
- optional real AI provider hanya melalui environment secret.

## 21.2 Staging/Demo

- database terpisah dari local;
- object storage private;
- fake provider untuk demo deterministic atau real provider dengan quota kecil;
- seeded dataset;
- fixture/simulation badge yang jelas;
- migration dijalankan melalui release step;
- log dan metrics dapat diperiksa.

## 21.3 Production Baseline

- API dan worker deployment terpisah;
- managed PostgreSQL dengan PostGIS;
- private object storage;
- secret manager/environment secret;
- TLS;
- backup dan tested restore;
- health/metrics/alerting;
- resource limits;
- deployment rollback plan.

## 21.4 Scaling Path

1. **Stage 1:** satu API instance, satu worker process, PostgreSQL queue.
2. **Stage 2:** tambah API replicas; worker tetap terpisah.
3. **Stage 3:** scale worker concurrency berdasarkan queue age/provider quota.
4. **Stage 4:** pindahkan job delivery ke broker eksternal jika database queue menjadi bottleneck yang terukur.
5. **Stage 5:** pisahkan AI worker atau media preprocessing hanya bila workload dan deployment cadence berbeda secara nyata.
6. **Stage 6:** partition/archive large event tables atau optimasi read model bila query/retention menuntut.

Jangan menambah infrastruktur sebelum bottleneck dapat dibuktikan melalui metrics.

---

# 22. Operational Runbook Minimum

## 22.1 AI Provider Down

1. Pastikan API intake tetap ready.
2. Periksa provider health/error rate.
3. Biarkan job menunggu atau retry sesuai policy.
4. Jangan memaksa retry massal.
5. Aktifkan circuit breaker bila threshold tercapai.
6. Setelah provider pulih, proses backlog dengan concurrency terbatas.

## 22.2 Queue Backlog

1. Periksa oldest queued age.
2. Periksa worker heartbeat dan resource.
3. Periksa rate limit provider.
4. Periksa apakah job stuck pada lease.
5. Tambah worker secara bertahap jika database/provider mampu.
6. Jangan reset seluruh job secara manual tanpa audit.

## 22.3 Deep Analysis Status Unknown

1. Periksa provider request/correlation ID.
2. Cari usage atau idempotency result dari provider bila tersedia.
3. Jangan membuat generation kedua sampai outcome pertama dipahami atau policy manual override digunakan.
4. Catat keputusan reconciliation.

## 22.4 False Duplicate

1. Buka relation detail.
2. Periksa image, koordinat, spatial distance, dan pHash distance.
3. Reverse relation dengan reason.
4. Kembalikan report ke workflow yang benar.
5. Reconcile auto-upvote berdasarkan relation event.
6. Pastikan audit tercatat.

## 22.5 Archive Salah

1. Periksa `completed_at`, `archived_at`, dan audit event.
2. Jangan mengubah database langsung tanpa prosedur.
3. Gunakan administrative correction jika disediakan.
4. Catat reason dan aktor.
5. Periksa scheduler cutoff/timezone configuration.

---

# 23. Definition of Done V1

V1 baru dapat disebut usable ketika seluruh kelompok berikut terpenuhi.

## Functional

- [ ] User dapat submit report dan media.
- [ ] Intake tidak memanggil AI.
- [ ] Report dan media tersimpan konsisten.
- [ ] Duplicate detection berjalan melalui PostGIS + pHash.
- [ ] Duplicate evidence tetap tersimpan.
- [ ] Worker triage berjalan asynchronous.
- [ ] Admin dapat melihat queue, batch, retry, dan failure.
- [ ] Deep analysis dapat diminta secara explicit.
- [ ] Deep analysis locked disajikan dari database tanpa provider call ulang.
- [ ] SPK berstatus draft sampai di-approve.
- [ ] Lifecycle transition tervalidasi.
- [ ] H+1 archiver menggunakan `completed_at`.
- [ ] Archived report read-only pada API normal.

## Reliability

- [ ] Job survive process restart.
- [ ] Retry tidak menggandakan vote atau side effect.
- [ ] Concurrent analysis tidak membuat double generation.
- [ ] Concurrent dedupe mempunyai protection.
- [ ] Worker heartbeat dapat dipantau.
- [ ] Provider outage tidak menghentikan intake.
- [ ] Archive sweep aman dijalankan berulang.

## Security

- [ ] Admin endpoint memerlukan auth.
- [ ] Permission dicek di backend.
- [ ] Tidak ada public AI generation endpoint.
- [ ] File media private dan diakses melalui authorization.
- [ ] Secret tidak masuk frontend/log/repository.
- [ ] Upload validation dan rate limit aktif.
- [ ] Admin actions penting mempunyai audit log.

## Testing

- [ ] Unit tests untuk state, hash, policy, dan cache.
- [ ] Integration tests PostgreSQL/PostGIS.
- [ ] Provider fake tests.
- [ ] Concurrency tests.
- [ ] API/E2E happy path.
- [ ] Failure path tests.
- [ ] Race detector untuk bagian Go yang relevan.

## Demo

- [ ] Dataset demo deterministic.
- [ ] Fixture dibedakan dari runtime metrics.
- [ ] Demo provider invocation count terlihat.
- [ ] Cache hit dapat dibuktikan.
- [ ] Duplicate decision dapat dijelaskan.
- [ ] H+1 dapat didemonstrasikan tanpa mengubah waktu production.
- [ ] Diagram dan API documentation tersedia.

---

# 24. Extension Points setelah V1

Desain V1 harus membuka extension point, tetapi tidak perlu mengimplementasikan semuanya sekarang.

## 24.1 AI Provider Registry

Tambahkan provider/model baru melalui adapter dan configuration. Business use case tetap menggunakan contract triage/deep analysis yang sama.

## 24.2 Prompt Experimentation

Tambahkan prompt version, evaluation dataset, output quality score, dan perbandingan model. Prompt baru tidak otomatis mengganti output locked yang sudah diterbitkan.

## 24.3 Advanced Duplicate Detection

Kemungkinan pengembangan:

- multi-image similarity;
- image embedding;
- text embedding;
- category-aware threshold;
- time decay;
- spatial density clustering.

Semua model similarity baru harus berjalan sebagai candidate signal terlebih dahulu dan dievaluasi sebelum boleh auto-merge.

## 24.4 Tenant Isolation

Jika FIXBACK digunakan banyak gedung/organisasi:

- wajibkan tenant scope pada query;
- tenant_id masuk ke unique constraints yang tepat;
- pisahkan role per tenant;
- pastikan object storage path dan signed access tenant-aware;
- evaluasi row-level security bila kompleksitas mengharuskannya.

## 24.5 Notification dan WhatsApp

Tambahkan notification module terpisah dengan:

- channel adapter;
- message template version;
- recipient consent;
- delivery status;
- provider request ID;
- retry/dead-letter;
- audit dan opt-out.

SPK export tidak boleh otomatis dianggap telah terkirim.

## 24.6 Analytics dan Reporting

Setelah data operasional stabil, tambahkan:

- trend kerusakan per lokasi;
- average resolution time;
- repeat incident rate;
- cost per triage;
- AI cache hit ratio;
- material estimate versus actual cost;
- facility risk heatmap.

Pisahkan operational database query dari analytical workload bila volume telah membenarkan read replica atau warehouse.

---

# 25. Glossary

| Term | Penjelasan |
|---|---|
| Admin-centric AI | AI generation dikendalikan backend/admin, bukan dipanggil langsung oleh end-user |
| Async processing | Pekerjaan berjalan setelah HTTP request selesai dan hasilnya dapat diperiksa kemudian |
| Modular monolith | Satu aplikasi/deployment codebase dengan batas modul internal yang jelas |
| PostGIS | Ekstensi PostgreSQL untuk menyimpan dan mencari data geospasial |
| Geography | Representasi spasial yang mendukung perhitungan jarak berbasis bumi, seperti meter |
| pHash | Perceptual hash yang meringkas karakter visual gambar |
| Hamming Distance | Jumlah bit berbeda antara dua hash |
| Idempotency | Operasi yang diulang dengan identitas request sama tidak menggandakan efek bisnis |
| Lease | Hak sementara worker untuk mengerjakan job sampai waktu kedaluwarsa |
| Dead letter | Job yang gagal melewati aturan retry dan memerlukan inspeksi/manual action |
| Fingerprint | Identitas ringkas dari input dan versi konfigurasi yang menentukan apakah hasil cache masih cocok |
| Cache hit | Hasil tersimpan digunakan kembali tanpa memanggil provider AI |
| Immutable version | Snapshot hasil yang tidak diedit; perubahan dibuat sebagai versi baru |
| Outbox pattern | Cara memastikan perubahan database dan niat mengirim job/event tercatat secara konsisten |
| RBAC | Role-Based Access Control, kontrol akses berdasarkan role dan permission |
| H+1 | Dalam blueprint ini berarti 24 jam penuh setelah `completed_at` |
| SPK | Surat Perintah Kerja; pada V1 hasil AI adalah draft sampai disetujui pihak berwenang |
| Observability | Kemampuan memahami kondisi sistem dari log, metrics, traces, dan health signals |

---

# 26. Rekomendasi Urutan Eksekusi Paling Efisien

Jika kapasitas pengerjaan terbatas, prioritaskan urutan berikut dan jangan mengerjakan visual polish sebelum alur data terbukti:

1. **Phase 0–3:** kontrak domain, database, dan ingestion. Pastikan report tersimpan tanpa AI.
2. **Phase 5–6:** job queue dan triage. Ini membuktikan asynchronous AI architecture.
3. **Phase 4:** duplicate detection. Integrasikan sebelum triage agar duplicate canonical dapat menghindari biaya AI.
4. **Phase 7–8:** admin API, Overview, dan Triage Hub.
5. **Phase 10:** deep analysis cache dan SPK draft. Ini bukti utama cost-control AI.
6. **Phase 11:** lifecycle dan H+1 archiver.
7. **Phase 9:** map/cluster UI, setelah backend duplicate decision stabil.
8. **Phase 12–13:** security hardening, test, demo dataset, dan release.

**Prioritas absolut:** intake yang tidak bergantung pada AI, job yang durable, idempotency, cache guard di backend, lifecycle transition yang tervalidasi, dan audit. Lima hal ini lebih penting daripada menambah banyak kartu dashboard atau animasi real-time.

---

**End of Blueprint — FIXBACK Admin Portal V1**
