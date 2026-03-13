# 安全文件盒

[**English**](README.md)

一个基于 Go + Gin 的 Web 应用，用于用户身份验证和加密文件存储，前端采用静态 HTML/CSS/JS。

**主要功能**
- 使用 JWT 进行用户注册/登录（含资料、头像、改密）
- 加密文件上传/下载（AES-256-GCM，分块）
- 批量上传 + 断点续传
- 文件预览（图片/文本/PDF/Office）
- 文件类型校验 + 可选的恶意软件扫描
- 由后端提供的静态 Web UI

---

## 1. 项目布局

- `cmd/server/main.go`：应用入口
- `internal/config/`：配置加载与验证
- `internal/handler/`：Gin HTTP 处理
- `internal/service/`：业务逻辑（文件加密在此）
- `internal/model/`：GORM 模型
- `internal/pkg/`：数据库、日志、辅助函数
- `internal/routes/`：API + 静态路由
- `web/templates/`：HTML 页面
- `web/static/`：JS/CSS/图片
- `storage/`：加密文件存储（运行时创建）
- `config.yaml`：运行时配置

---

## 2. 前提条件

- Go 1.18+（建议与 `go.mod` 版本匹配）
- MySQL 8+（或兼容版本）
- 可选：ClamAV（`clamscan` 或 `clamdscan`）用于恶意软件扫描
- 可选：LibreOffice（`libreoffice` 或 `soffice`）用于 Office 预览

---

## 3. 配置（`config.yaml`）

必需字段：

- `database.*`：数据库连接参数
- `jwt.secret`：JWT 签名密钥（至少 32 个字符）
- `jwt.expiry_minutes`：JWT 过期时间（分钟）
- `file_crypto.key`：**base64 URL 安全**密钥（解码后至少 32 字节）
- `malware_scan.*`：可选的恶意软件扫描配置

示例（仓库内已有同名文件）：

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
  secret: <您的强密钥>

file_crypto:
  key: <base64-url-encoded-32-bytes>

malware_scan:
  enabled: true
  command: ""          # 为空时自动探测 clamscan/clamdscan
  timeout_seconds: 30
  allow_on_failure: false
```

备注：
- 启动时，如果 `jwt.secret` 或 `file_crypto.key` 缺失或强度不足，应用程序会**自动**生成并写回 `config.yaml`。
- `file_crypto.key` 必须是 base64 URL 安全密钥（无填充）。示例生成：
- 也可通过环境变量覆盖，例如 `JWT_SECRET`、`FILE_CRYPTO_KEY`。

```bash
python - <<'PY'
import os, base64
print(base64.urlsafe_b64encode(os.urandom(32)).rstrip(b'=').decode())
PY
```

---

## 4. 数据库设置

创建数据库（名称需与 `config.yaml` 一致）：

```sql
CREATE DATABASE secure_file_box;
```

设置 MySQL root 密码与 `config.yaml` 匹配（示例）：

```sql
ALTER USER 'root'@'localhost' IDENTIFIED BY 'yourpassword';
```

---

## 5. 运行（开发）

在仓库根目录执行：

```bash
go run ./cmd/server/main.go
```

打开：
- `http://127.0.0.1:8080`

---

## 6. 构建（生产）

```bash
go build -o ./bin/app ./cmd/server
./bin/app
```

---

## 7. API 概览

所有 API 均挂载在 `/api/v1`。

- `GET /api/v1/ping`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/user/profile`
- `PUT /api/v1/user/profile`
- `GET /api/v1/user/avatar`
- `PUT /api/v1/user/avatar`
- `PUT /api/v1/user/password`

文件：
- `POST /api/v1/files/upload`（需要 JWT）
- `POST /api/v1/files/batch`（需要 JWT）
- `POST /api/v1/files/public/upload`（无需 JWT）
- `GET /api/v1/files`（需要 JWT）
- `GET /api/v1/files/download/:id`（需要 JWT）
- `GET /api/v1/files/preview/:id`（需要 JWT）
- `PUT /api/v1/files/:id`（需要 JWT）
- `DELETE /api/v1/files/:id`（需要 JWT）
- `DELETE /api/v1/files/batch`（需要 JWT）
- `POST /api/v1/files/resumable/init`（需要 JWT）
- `GET /api/v1/files/resumable/:upload_id`（需要 JWT）
- `POST /api/v1/files/resumable/:upload_id/chunk`（需要 JWT）
- `POST /api/v1/files/resumable/:upload_id/complete`（需要 JWT）
- `DELETE /api/v1/files/resumable/:upload_id`（需要 JWT）

同时保留无 `/api/v1` 前缀的旧路由，便于兼容旧客户端。

---

## 8. 上传与校验

- 允许的扩展名：`jpg`, `jpeg`, `png`, `gif`, `webp`, `txt`, `md`, `json`, `log`, `csv`, `pdf`, `doc`, `docx`, `xls`, `xlsx`, `ppt`, `pptx`。
- 禁止的扩展名：`exe`, `dll`, `so`, `bin`, `sh`, `bat`, `apk`, `dmg`, `iso`, `msi`, `com`, `scr`。
- 内容校验会检查 MIME 类型与文件头；文本文件必须是 UTF-8。
- 断点续传会将 `chunk_size` 限制在 256 KB–20 MB。
- 恶意软件扫描由 `malware_scan.*` 控制并使用 ClamAV。开启扫描且无可用扫描器时，除非 `allow_on_failure: true`，否则上传会失败。

---

## 9. 预览

- `/files/preview/:id` 支持图片、文本、PDF 与 Office 文件预览。
- 文本预览最多 2 MB，编码尝试 UTF-8/UTF-16/GB18030/ISO-8859-1。
- Office 预览需要安装 LibreOffice（`libreoffice` 或 `soffice`）进行在线转换。

---

## 10. 加密详情

本项目的文件与元数据均使用 AES-256-GCM 进行认证加密，密钥由 `file_crypto.key` 派生。

**密钥策略**
- `file_crypto.key` 必须是 Base64 URL-safe（无填充）的密钥，解码后长度至少 32 字节。
- 使用 HMAC-SHA256 从同一主密钥派生两把子密钥：
- 文件内容密钥：`HMAC(key, "file-gcm-aes256")`
- 元数据密钥：`HMAC(key, "db-meta-gcm-aes256")`

**文件加密（分块）**
- 算法：AES-256-GCM。
- 分块大小：32 KB。
- 文件头格式：`SFB2` 魔数 + 8 字节随机前缀（nonce prefix）。
- 每个分块的 nonce：`prefix(8)` + `counter(4)`（大端递增计数）。
- 附加认证数据（AAD）：4 字节计数器（大端）。
- 每个分块存储格式：`uint32(len(sealed))`（大端）+ `sealed`（密文 + GCM tag）。
- 解密时逐块认证并写出，任一分块认证失败即返回 `file integrity check failed`。

**元数据加密（数据库字段）**
- 字段：文件名、存储路径、大小、描述、上传者 ID、MIME。
- 每个字段独立加密，随机 12 字节 nonce。
- 存储格式：`v1:` + Base64 URL-safe（无填充）编码的 `nonce || sealed`。
- 解密失败将报 `metadata integrity check failed`，列表接口会跳过该条记录以避免影响整体返回。

**兼容与迁移**
- 若数据库中 `enc_*` 字段为空，会回退读取旧字段（`legacy_*`），用于兼容旧数据。

**重要提示**
- 更改 `file_crypto.key` 会导致已有文件与元数据无法解密。
- 若看到 `invalid file magic` 或 `invalid encrypted metadata format`，通常是密钥不匹配、格式变更或数据损坏。

---

## 11. 测试

测试位于 `test/` 目录，使用临时 SQLite 数据库（无需 MySQL）。测试中默认关闭恶意软件扫描，Office 预览转换未覆盖。

运行全部测试：

```bash
go test ./...
```

仅运行测试包：

```bash
go test ./test -v
```

覆盖点：
- `test/config_test.go`：密钥生成与配置写回。
- `test/file_validation_test.go`：扩展名允许/阻止与内容校验。
- `test/file_service_test.go`：加解密流程、限流、旧字段兼容。
- `test/resumable_upload_test.go`：断点续传流程与异常场景。
- `test/user_service_test.go`：用户创建/认证/改密/资料更新。
- `test/jwt_middleware_test.go`：JWT 中间件正常与未授权路径。
- `test/utils_test.go`：密码哈希与分页默认值。

备注：
- 文件加密使用 `test/test_helpers.go` 中固定的测试密钥。
- 临时文件与 SQLite 数据库均使用 `t.TempDir()` 自动清理。

---

## 12. 故障排除

- **MySQL 身份验证错误**：检查 `database.user/password` 是否正确，以及数据库是否可访问。
- **文件魔数无效/完整性检查失败**：文件被不同的 `file_crypto.key` 加密、使用旧格式或已损坏。
- **启动时密钥错误**：确保 `file_crypto.key` 是有效的 base64 URL 安全密钥，且解码后至少 32 字节。
- **预览不可用**：安装 LibreOffice（`libreoffice`/`soffice`）。
- **恶意软件扫描失败**：检查 `malware_scan.command` 或安装 ClamAV。

---

## 13. 部署说明

- 生产环境建议使用环境变量或密钥管理器。
- 在 Go 服务前配置 Nginx/Traefik 以启用 TLS。
- 备份 `storage/` 与数据库。

---

## 14. 贡献

在进行重大更改前请先提交 issue。尽量减少改动，并尽可能包含测试。
