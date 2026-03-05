# Secure File Box
[**简体中文**](README_zh_CN.md)

A Go + Gin web app for user auth and encrypted file storage with a static HTML/CSS/JS frontend.

**Key features**
- User registration/login with JWT
- Encrypted file upload/download (AES-256-GCM)
- File listing & deletion
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

---

## 3. Configuration (`config.yaml`)

Minimal required fields:

- `database.*`: DB connection parameters
- `jwt.secret`: JWT signing secret (min 32 chars)
- `file_crypto.key`: **base64 url-safe** secret (min 32 bytes after decoding)

Example (already in repo):

```yaml
server:
  host: 127.0.0.1
  port: 8080

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
  secret: <your-strong-secret>

file_crypto:
  key: <base64-url-encoded-32-bytes>
```

Notes:
- On startup, if `jwt.secret` or `file_crypto.key` is missing/weak, the app **auto-generates** and writes it back to `config.yaml`.
- `file_crypto.key` must be base64 URL-safe (no padding). Example generation:

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

- `POST /api/v1/auth/signup`
- `POST /api/v1/auth/login`
- `GET /api/v1/user/profile`
- `PUT /api/v1/user/profile`

Files:
- `POST /api/v1/files/upload` (JWT required)
- `POST /api/v1/files/public/upload` (no JWT)
- `GET /api/v1/files` (JWT required)
- `GET /api/v1/files/download/:id` (JWT required)
- `DELETE /api/v1/files/:id` (JWT required)

---

## 8. Encryption Details

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
- Fields: filename, storage path, size, description, uploader ID.
- Each field is encrypted independently with a random 12-byte nonce.
- Stored format: `v1:` + Base64 URL-safe (no padding) of `nonce || sealed`.
- Decrypt failures return `metadata integrity check failed`; list API skips such rows to avoid breaking the entire response.

**Compatibility and migration**
- If `enc_*` fields are empty, the service falls back to legacy fields (`legacy_*`).

**Important**
- Changing `file_crypto.key` will make existing files and metadata unreadable.
- `invalid file magic` or `invalid encrypted metadata format` usually means key mismatch, format change, or corruption.

---

## 9. Testing

No automated tests are included yet.

---

## 10. Troubleshooting

- **MySQL auth error**: verify `database.user/password` and DB is reachable.
- **Invalid file magic / integrity check failed**: file was encrypted with a different `file_crypto.key`, uses an old format, or is corrupted.
- **Key errors at startup**: ensure `file_crypto.key` is valid base64 URL-safe and decodes to at least 32 bytes.

---

## 11. Deployment Notes

- Use environment variables or secret manager in production.
- Put Nginx/Traefik in front of the Go server for TLS.
- Back up `storage/` and DB together.

---

## 12. Contributing

Open an issue before large changes. Keep changes small and include tests where possible.
