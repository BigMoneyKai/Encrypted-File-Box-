# Secure File Box
[**简体中文**](README_zh_CN.md)

A Go + Gin web app for user auth and encrypted file storage with a static HTML/CSS/JS frontend.

**Key features**
- User registration/login with JWT (+ profile, avatar, password change)
- Encrypted file upload/download (AES-256-GCM, chunked)
- Batch upload + resumable uploads
- File preview (images/text/pdf/office)
- File type validation + optional malware scan
- Static web UI served by the backend

---

## 1. Project Layout

- `cmd/server/main.go`: app entrypoint
- `internal/config/`: config loading and validation
- `internal/handler/`: Gin HTTP handlers
- `internal/service/`: business logic (file encryption lives here)
- `internal/model/`: GORM models
- `internal/pkg/`: DB, logger, helpers
- `internal/routes/`: API + static routes
- `web/templates/`: HTML pages
- `web/static/`: JS/CSS/images
- `storage/`: encrypted file blobs (created at runtime)
- `config.yaml`: runtime configuration

---

## 2. Prerequisites

- Go 1.18+ (recommended to match `go.mod`)
- MySQL 8+ (or compatible)
- Optional: ClamAV (`clamscan` or `clamdscan`) for malware scanning
- Optional: LibreOffice (`libreoffice` or `soffice`) for Office preview

---

## 3. Configuration (`config.yaml`)

Minimal required fields:

- `database.*`: DB connection parameters
- `jwt.secret`: JWT signing secret (min 32 chars)
- `jwt.expiry_minutes`: token lifetime in minutes
- `file_crypto.key`: **base64 url-safe** secret (min 32 bytes after decoding)
- `malware_scan.*`: optional malware scan behavior

Example (already in repo):

```yaml
server:
  app_name: secure_file_box
  env: development
  debug: true
  host: 127.0.0.1
  port: 8080
  time_zone: Asia/Shanghai

database:
  driver: mysql
  host: localhost
  port: 3306
  user: root
  password: "0827"
  name: secure_file_box

jwt:
  issuer: secure_file_box
  audience: secure_users
  expiry_minutes: 60
  secret: <your-strong-secret>

file_crypto:
  key: <base64-url-encoded-32-bytes>

malware_scan:
  enabled: true
  command: ""          # empty = auto-detect clamscan/clamdscan
  timeout_seconds: 30
  allow_on_failure: false
```

Notes:
- On startup, if `jwt.secret` or `file_crypto.key` is missing/weak, the app **auto-generates** and writes it back to `config.yaml`.
- `file_crypto.key` must be base64 URL-safe (no padding). Example generation:
- Config can also be overridden by environment variables like `JWT_SECRET` and `FILE_CRYPTO_KEY`.

```bash
python - <<'PY'
import os, base64
print(base64.urlsafe_b64encode(os.urandom(32)).rstrip(b'=').decode())
PY
```

---

## 4. Database Setup

Create the database (schema name must match `config.yaml`):

```sql
CREATE DATABASE secure_file_box;
```

Set MySQL root password to match your `config.yaml` (example):

```sql
ALTER USER 'root'@'localhost' IDENTIFIED BY 'yourpassword';
```

---

## 5. Run (Dev)

From repo root:

```bash
go run ./cmd/server/main.go
```

Open:

- `http://127.0.0.1:8080`

---

## 6. Build (Prod)

```bash
go build -o ./bin/app ./cmd/server
./bin/app
```

---

## 7. API Overview

All APIs are mounted under `/api/v1`.

- `GET /api/v1/ping`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/user/profile`
- `PUT /api/v1/user/profile`
- `GET /api/v1/user/avatar`
- `PUT /api/v1/user/avatar`
- `PUT /api/v1/user/password`

Files:
- `POST /api/v1/files/upload` (JWT required)
- `POST /api/v1/files/batch` (JWT required)
- `POST /api/v1/files/public/upload` (no JWT)
- `GET /api/v1/files` (JWT required)
- `GET /api/v1/files/download/:id` (JWT required)
- `GET /api/v1/files/preview/:id` (JWT required)
- `PUT /api/v1/files/:id` (JWT required)
- `DELETE /api/v1/files/:id` (JWT required)
- `DELETE /api/v1/files/batch` (JWT required)
- `POST /api/v1/files/resumable/init` (JWT required)
- `GET /api/v1/files/resumable/:upload_id` (JWT required)
- `POST /api/v1/files/resumable/:upload_id/chunk` (JWT required)
- `POST /api/v1/files/resumable/:upload_id/complete` (JWT required)
- `DELETE /api/v1/files/resumable/:upload_id` (JWT required)

Legacy routes (no `/api/v1` prefix) are also available for older clients.

---

## 8. Uploads & Validation

- Allowed extensions: `jpg`, `jpeg`, `png`, `gif`, `webp`, `txt`, `md`, `json`, `log`, `csv`, `pdf`, `doc`, `docx`, `xls`, `xlsx`, `ppt`, `pptx`.
- Blocked extensions: `exe`, `dll`, `so`, `bin`, `sh`, `bat`, `apk`, `dmg`, `iso`, `msi`, `com`, `scr`.
- Content validation checks MIME type and magic bytes; text files must be UTF-8.
- Resumable uploads clamp `chunk_size` to 256 KB–20 MB.
- Malware scanning is controlled by `malware_scan.*` and uses ClamAV. If scanning is enabled and no scanner is available, uploads fail unless `allow_on_failure: true`.

---

## 9. Preview

- `/files/preview/:id` supports images, text, PDF, and Office files.
- Text previews are limited to 2 MB and decoded as UTF-8/UTF-16/GB18030/ISO-8859-1.
- Office previews require LibreOffice (`libreoffice` or `soffice`) to convert to PDF on the fly.

---

## 10. Encryption Details

File content and metadata are both protected with AES-256-GCM, with keys derived from `file_crypto.key`.

**Key strategy**
- `file_crypto.key` must be Base64 URL-safe (no padding) and decode to at least 32 bytes.
- Two subkeys are derived via HMAC-SHA256 from the same master key:
- File content key: `HMAC(key, "file-gcm-aes256")`
- Metadata key: `HMAC(key, "db-meta-gcm-aes256")`

**File encryption (chunked)**
- Algorithm: AES-256-GCM.
- Chunk size: 32 KB.
- File header: magic `SFB2` + 8-byte random nonce prefix.
- Per-chunk nonce: `prefix(8)` + `counter(4)` (big-endian, increasing).
- AAD: 4-byte counter (big-endian).
- Chunk storage format: `uint32(len(sealed))` (big-endian) + `sealed` (ciphertext + GCM tag).
- Decryption authenticates each chunk; any failure returns `file integrity check failed`.

**Metadata encryption (DB fields)**
- Fields: filename, storage path, size, description, uploader ID, MIME.
- Each field is encrypted independently with a random 12-byte nonce.
- Stored format: `v1:` + Base64 URL-safe (no padding) of `nonce || sealed`.
- Decrypt failures return `metadata integrity check failed`; list API skips such rows to avoid breaking the entire response.

**Compatibility and migration**
- If `enc_*` fields are empty, the service falls back to legacy fields (`legacy_*`).

**Important**
- Changing `file_crypto.key` will make existing files and metadata unreadable.
- `invalid file magic` or `invalid encrypted metadata format` usually means key mismatch, format change, or corruption.

---

## 11. Testing

Tests live under `test/` and use a temporary SQLite database (no MySQL required). Malware scanning is disabled in tests, and Office preview conversion is not exercised.

Run all tests:

```bash
go test ./...
```

Run only the test package:

```bash
go test ./test -v
```

Key coverage:
- `test/config_test.go`: secret/key generation and config write-back.
- `test/file_validation_test.go`: extension allow/deny and content validation.
- `test/file_service_test.go`: encrypt/decrypt flow, limits, legacy metadata fallback.
- `test/resumable_upload_test.go`: init/chunk/complete flow and error cases.
- `test/user_service_test.go`: user create/auth/password/profile flows.
- `test/jwt_middleware_test.go`: JWT middleware happy/unauthorized paths.
- `test/utils_test.go`: password hashing and pagination defaults.

Notes:
- File encryption uses a fixed test key via `test/test_helpers.go`.
- Temp files and sqlite DBs are created under `t.TempDir()` and cleaned automatically.

---

## 12. Troubleshooting

- **MySQL auth error**: verify `database.user/password` and DB is reachable.
- **Invalid file magic / integrity check failed**: file was encrypted with a different `file_crypto.key`, uses an old format, or is corrupted.
- **Key errors at startup**: ensure `file_crypto.key` is valid base64 URL-safe and decodes to at least 32 bytes.
- **Preview unavailable**: install LibreOffice (`libreoffice`/`soffice`).
- **Malware scan failed**: verify `malware_scan.command` or install ClamAV.

---

## 13. Deployment Notes

- Use environment variables or secret manager in production.
- Put Nginx/Traefik in front of the Go server for TLS.
- Back up `storage/` and DB together.

---

## 14. Contributing

Open an issue before large changes. Keep changes small and include tests where possible.
