# File Manager

Production-oriented file management service built with Go, gRPC, REST gateway, MySQL, RabbitMQ, ClamAV, and S3-compatible object storage.

The service keeps application servers out of the file transfer path. Clients request pre-signed URLs, upload files directly to object storage, complete the upload through the API, and then a background worker scans the uploaded object for viruses by streaming it to ClamAV.

## Highlights

- Direct-to-storage uploads with S3 pre-signed URLs
- Multipart upload orchestration for large files
- gRPC API with HTTP/JSON gateway
- MySQL-backed file metadata and lifecycle state
- RabbitMQ event flow for asynchronous virus scanning
- ClamAV integration behind a scanner abstraction
- S3-compatible storage abstraction with MinIO support
- Structured logging with Zap
- Clean internal layering: handler, service, repository, storage, messaging, scanner

## Architecture

```text
                            Client
                              |
                              | HTTP/JSON or gRPC
                              v
                     File Manager API
                              |
          +-------------------+-------------------+
          |                                       |
          v                                       v
       MySQL                              S3-compatible storage
 file metadata                         MinIO / Amazon S3 / R2
          |
          | complete_upload event
          v
       RabbitMQ
          |
          v
 Virus Scanner Worker
          |
          | stream object body
          v
        ClamAV
```

### Upload Flow

1. Client requests an upload URL from the API.
2. API validates file metadata and creates a pending file record.
3. Client uploads the file directly to object storage.
4. Client calls the complete endpoint.
5. API verifies the object exists, marks the file completed, and publishes a `complete_upload` event.
6. Virus scanner worker consumes the event and streams the object to ClamAV.
7. Worker updates the file virus scan status to clean, infected, or failed.

## Services

### API

`cmd/api/main.go`

Runs the public gRPC server and HTTP gateway. It owns request validation, file metadata workflows, pre-signed URL generation, and upload completion events.

### Virus Scanner Worker

`cmd/virus-scanner-worker/main.go`

Consumes `complete_upload` messages from RabbitMQ and calls `ScanFile()` in the service layer. The worker does not download files to disk. It opens an object storage stream and passes it into the scanner abstraction.

## Core Packages

```text
cmd/
  api/                         API process
  virus-scanner-worker/        Background virus scan worker

internal/
  config/                      Environment-based configuration
  consumer/                    RabbitMQ message handlers
  database/                    MySQL connection setup
  errs/                        Domain errors and gRPC mapping
  handler/                     gRPC handlers and request validation
  interceptor/                 gRPC auth and request logging interceptors
  messaging/                   Messaging abstraction and RabbitMQ adapter
  model/                       Domain models and statuses
  repository/                  File metadata repository abstraction
  repository/mysql/            MySQL repository implementation
  service/                     File business workflows
  storage/                     Object storage abstraction
  storage/s3/                  S3-compatible storage implementation
  virus-scanner/               Virus scanner abstraction
  virus-scanner/clamav/        ClamAV implementation

proto/
  file/v1/                     API contract

gen/
  go/                          Generated Go protobuf and gateway files
  openapi/                     Generated OpenAPI specification

migrations/                    Database schema migrations
```

## Technology Stack

- Go 1.25+
- gRPC and Protocol Buffers
- gRPC-Gateway for HTTP/JSON
- MySQL 8
- RabbitMQ
- ClamAV
- AWS SDK for Go v2
- S3-compatible object storage, including MinIO
- GORM
- Docker Compose
- Buf
- Zap

## API Surface

All HTTP endpoints are generated from `proto/file/v1/file.proto`.

Requests must include `x-user-id`. The current interceptor uses this header as the caller identity and injects it into request context.

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/v1/file/upload-url` | Create a pre-signed URL for a simple upload |
| `POST` | `/v1/file/{file_id}/complete` | Complete a simple upload |
| `POST` | `/v1/file/multipart` | Create a multipart upload session |
| `POST` | `/v1/file/multipart/url` | Create a pre-signed URL for one multipart part |
| `POST` | `/v1/file/multipart/complete` | Complete a multipart upload |
| `POST` | `/v1/file/multipart/abort` | Abort a multipart upload |
| `GET` | `/v1/file/{file_id}` | Get file metadata |
| `POST` | `/v1/file/{file_id}/download-url` | Create a temporary download URL |
| `DELETE` | `/v1/file/{file_id}` | Delete file metadata and object |

## Local Development

### Prerequisites

- Go 1.25+
- Docker and Docker Compose
- Buf CLI
- golang-migrate CLI

### Start Infrastructure

```bash
docker compose up -d
```

Docker Compose starts:

- MySQL on `localhost:3306`
- MinIO API on `localhost:9000`
- MinIO console on `localhost:9001`
- RabbitMQ on `localhost:5672`
- RabbitMQ management UI on `localhost:15672`
- ClamAV on `localhost:3310`

Create the object storage bucket used by the service. With the default values below, create a bucket named `file-manager` in MinIO.

### Environment

Create a `.env` file in the project root:

```env
APP_NAME=file-manager
APP_ENV=development
APP_HTTP_PORT=8080
APP_GRPC_PORT=9090
APP_DEBUG=true

DB_HOST=localhost
DB_PORT=3306
DB_USER=mysqladmin
DB_PASSWORD=mysqladmin
DB_NAME=file_manager
DB_TLS=skip-verify

S3_ENDPOINT=http://localhost:9000
S3_REGION=us-east-1
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=file-manager
S3_USE_SSL=false

PRESIGNED_UPLOAD_EXPIRE_MINUTES=15m
PRESIGNED_DOWNLOAD_EXPIRE_MINUTES=30m
MULTIPART_PART_SIZE=5242880

MAX_FILE_SIZE=104857600
ALLOWED_FILE_TYPES=image/jpeg,image/png,application/pdf

RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=rabbitmq
RABBITMQ_PASSWORD=rabbitmq
RABBITMQ_VHOST=/

CLAMAV_ENABLED=true
CLAMAV_HOST=localhost
CLAMAV_PORT=3310
CLAMAV_NETWORK=tcp
CLAMAV_CHUNK_SIZE=1048576
CLAMAV_TIMEOUT=5m

RATE_LIMIT_REQUEST=100
RATE_LIMIT_DURATION=1m
```

### Run Database Migrations

```bash
migrate -path ./migrations \
  -database "mysql://mysqladmin:mysqladmin@tcp(localhost:3306)/file_manager" \
  up
```

### Generate API Code

Run this after changing files under `proto/`:

```bash
buf generate
```

### Run The API

```bash
go run ./cmd/api/main.go
```

The HTTP gateway listens on `http://localhost:8080` by default. The gRPC server listens on `localhost:9090`.

### Run The Virus Scanner Worker

```bash
go run ./cmd/virus-scanner-worker/main.go
```

The worker requires RabbitMQ, object storage, MySQL, and ClamAV to be available.

## Example Simple Upload

### 1. Request Upload URL

```bash
curl -X POST http://localhost:8080/v1/file/upload-url \
  -H "content-type: application/json" \
  -H "x-user-id: user-123" \
  -d '{
    "file_name": "invoice.pdf",
    "content_type": "application/pdf",
    "size": 1048576
  }'
```

Example response:

```json
{
  "file_id": "2feb8bfb-fc8f-4674-a633-58174b70718f",
  "upload_url": "http://localhost:9000/file-manager/user-123/...",
  "object_key": "user-123/2feb8bfb-fc8f-4674-a633-58174b70718f.pdf",
  "headers": {}
}
```

### 2. Upload To Object Storage

```bash
curl -X PUT \
  -H "content-type: application/pdf" \
  --upload-file invoice.pdf \
  "PRESIGNED_UPLOAD_URL"
```

### 3. Complete Upload

```bash
curl -X POST http://localhost:8080/v1/file/2feb8bfb-fc8f-4674-a633-58174b70718f/complete \
  -H "content-type: application/json" \
  -H "x-user-id: user-123" \
  -d '{}'
```

### 4. Check Metadata

```bash
curl http://localhost:8080/v1/file/2feb8bfb-fc8f-4674-a633-58174b70718f \
  -H "x-user-id: user-123"
```

## Multipart Upload Flow

Use multipart upload for large files or clients that need resumable part uploads.

1. Call `POST /v1/file/multipart` with file metadata.
2. For each part, call `POST /v1/file/multipart/url`.
3. Upload each part directly to object storage using the returned URL.
4. Call `POST /v1/file/multipart/complete` with the uploaded part numbers and ETags.
5. The API publishes `complete_upload`, and the virus scanner worker processes the final object.

## Virus Scanning

Virus scanning is intentionally isolated behind `internal/virus-scanner.Scanner`.

The service layer depends only on the scanner interface, not on ClamAV directly. Replacing ClamAV with another engine should only require a new implementation under `internal/virus-scanner` and a different worker wiring.

The current ClamAV adapter uses stream scanning. The worker reads from object storage as an `io.Reader` and sends the stream to ClamAV without storing the full file in memory or on local disk.

Operational notes:

- ClamAV must be running and ready before starting the worker.
- ClamAV scan limits should be aligned with `MAX_FILE_SIZE`.
- Large scans still transfer the full object over the network because the scanner must inspect the file bytes.
- Files that cannot be scanned are marked as failed instead of being retried forever.

## Configuration Reference

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `APP_NAME` | No | `file-manager` | Application name |
| `APP_ENV` | No | `development` | Runtime environment |
| `APP_HTTP_PORT` | No | `8080` | HTTP gateway port |
| `APP_GRPC_PORT` | No | `9090` | gRPC server port |
| `APP_DEBUG` | No | `true` | Enables development logger |
| `DB_HOST` | Yes | - | MySQL host |
| `DB_PORT` | No | `3306` | MySQL port |
| `DB_USER` | Yes | - | MySQL user |
| `DB_PASSWORD` | No | - | MySQL password |
| `DB_NAME` | Yes | - | MySQL database |
| `DB_TLS` | No | `skip-verify` | MySQL TLS mode |
| `S3_ENDPOINT` | Yes | - | S3-compatible endpoint |
| `S3_ACCESS_KEY` | Yes | - | S3 access key |
| `S3_SECRET_KEY` | Yes | - | S3 secret key |
| `S3_BUCKET` | Yes | - | Object bucket |
| `S3_USE_SSL` | No | `false` | Enable HTTPS for S3 endpoint |
| `S3_REGION` | No | `us-east-1` | S3 region |
| `PRESIGNED_UPLOAD_EXPIRE_MINUTES` | No | `15m` | Upload URL TTL |
| `PRESIGNED_DOWNLOAD_EXPIRE_MINUTES` | No | `30m` | Download URL TTL |
| `MULTIPART_PART_SIZE` | No | `5242880` | Multipart part size in bytes |
| `MAX_FILE_SIZE` | No | `104857600` | Maximum accepted file size in bytes |
| `ALLOWED_FILE_TYPES` | No | - | Comma-separated content type allowlist |
| `RABBITMQ_HOST` | Yes | - | RabbitMQ host |
| `RABBITMQ_PORT` | Yes | - | RabbitMQ port |
| `RABBITMQ_USER` | Yes | - | RabbitMQ user |
| `RABBITMQ_PASSWORD` | Yes | - | RabbitMQ password |
| `RABBITMQ_VHOST` | Yes | - | RabbitMQ virtual host, usually `/` locally |
| `CLAMAV_ENABLED` | No | `false` | Enables virus scanner worker startup |
| `CLAMAV_HOST` | No | `localhost` | ClamAV host |
| `CLAMAV_PORT` | No | `3310` | ClamAV TCP port |
| `CLAMAV_NETWORK` | No | `tcp` | ClamAV network type |
| `CLAMAV_CHUNK_SIZE` | No | `1048576` | Scanner stream chunk size |
| `CLAMAV_TIMEOUT` | No | `5m` | Scanner timeout |
| `RATE_LIMIT_REQUEST` | No | `100` | Rate limit request count |
| `RATE_LIMIT_DURATION` | No | `1m` | Rate limit window |

## Production Notes

Before running this service in production, review these areas:

- Replace the development `x-user-id` identity mechanism with your production authentication and authorization layer.
- Use managed MySQL with backups, migrations, connection limits, and TLS.
- Use production-grade object storage credentials and bucket policies.
- Configure RabbitMQ durability, dead-letter queues, retry policy, and monitoring.
- Align ClamAV `StreamMaxLength`, `MaxFileSize`, and `MaxScanSize` with `MAX_FILE_SIZE`.
- Add health checks for API, worker, RabbitMQ, MySQL, object storage, and ClamAV.
- Export metrics and traces with your observability platform.
- Set strict presigned URL TTLs and content type allowlists.
- Run the API and worker as separate deployable processes.
- Add CI checks for tests, protobuf generation, migrations, linting, and container builds.

## Testing

Run all Go tests:

```bash
go test ./...
```

Run service tests only:

```bash
go test ./internal/service
```

## Troubleshooting

### RabbitMQ says `no access to this vhost`

Use the default vhost locally:

```env
RABBITMQ_VHOST=/
```

Grant permissions if needed:

```bash
docker exec file-manager-rabbitmq \
  rabbitmqctl set_permissions -p / rabbitmq ".*" ".*" ".*"
```

### ClamAV returns `broken pipe`

This usually means ClamAV closed the stream. Check whether the daemon is ready and whether scan limits are large enough:

```bash
docker logs file-manager-clamav --tail=200
printf 'zPING\0' | nc -w 3 localhost 3310
```

Expected response:

```text
PONG
```

### MinIO upload fails

Confirm the bucket exists and that `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, and `S3_SECRET_KEY` match your MinIO setup.

## License

MIT
