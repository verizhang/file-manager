# File Manager Service

Production-ready file manager service built with Go, gRPC, gRPC-Gateway, and S3-compatible object storage.

This project is designed to provide scalable and cloud-native file management capabilities with support for:

- Presigned upload/download URL
- Multipart upload
- Direct-to-storage upload architecture
- Preview URL generation
- gRPC + REST API
- Kubernetes-friendly deployment
- Virus scanning ready architecture

---

# Features

## Simple Upload

Generate presigned upload URLs for small file uploads.

- Direct upload to S3/MinIO
- Reduced API server bandwidth usage
- Optimized for cloud-native environments

---

## Multipart Upload

Designed for large file uploads.

- Multipart upload session
- Chunk upload support
- Parallel upload ready
- Retry-per-chunk architecture
- Large file upload support
- Resumable-ready design

---

## File Access

- Generate presigned download URL
- Generate preview URL
- Retrieve file metadata
- Soft delete support

---

## Infrastructure Ready

- Docker Compose local development
- Kubernetes-ready architecture
- Environment-based configuration
- Structured logging support
- External object storage support

---

# Architecture Overview

```text
Client
   │
   │ Request Upload URL
   ▼
File Manager Service
   │
   │ Generate Presigned URL
   ▼
S3 / MinIO
   ▲
   │ Direct Upload
   │
Client
```

### Why Presigned Upload?

The API server only handles upload orchestration and metadata.

Actual binary transfer happens directly between client and object storage.

Benefits:

- Lower API server bandwidth usage
- Better horizontal scalability
- Reduced infrastructure bottleneck
- Optimized for Kubernetes environments

---

# Tech Stack

| Area | Technology |
|---|---|
| Language | Go |
| API Protocol | gRPC |
| REST Gateway | gRPC-Gateway |
| Database | MySQL |
| ORM | GORM |
| Object Storage | MinIO / S3 Compatible |
| Containerization | Docker |
| Deployment | Kubernetes |
| Configuration | envconfig |
| Logging | Zap |

---

# Project Structure

```bash
.
├── cmd/
│   └── api/
│
├── internal/
│   ├── config/
│   ├── database/
│   ├── handler/
│   ├── logger/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   │   ├── entity/
│   │   └── mapper/
│   ├── server/
│   ├── service/
│   └── storage/
│
├── proto/
│   └── file/
│       └── v1/
│
├── gen/
│   ├── go/
│   └── openapi/
│
├── migrations/
│
├── docker-compose.yaml
├── buf.yaml
├── buf.gen.yaml
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

---

# Local Development

## Requirements

- Go 1.24+
- Docker
- Docker Compose

---

## Start Infrastructure

```bash
docker compose up -d
```

This will start:

- MinIO
- MySQL
- ClamAV

---

## Run Database Migration

This project intentionally uses raw SQL migration files instead of ORM auto-migration
to provide better schema control and production safety.

Migration files are located in:

```bash
migrations/
```

### Install golang-migrate

macOS:

```bash
brew install golang-migrate
```

Linux:

```bash
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz \
| tar xvz
sudo mv migrate /usr/local/bin/
```

---

### Run Migration

```bash
migrate \
  -path migrations \
  -database "mysql://mysqladmin:mysqladmin@tcp(localhost:3306)/file_manager" \
  up
```

---

### Rollback Migration

```bash
migrate \
  -path migrations \
  -database "mysql://mysqladmin:mysqladmin@tcp(localhost:3306)/file_manager" \
  down
```

---

## Generate Protobuf Files

```bash
buf generate
```

This will generate:

- gRPC code
- gRPC-Gateway code
- OpenAPI/Swagger specification

---

## Run Application

```bash
go run cmd/api/main.go
```

---

# Environment Variables

Copy example environment:

```bash
cp .env.example .env
```

Then adjust values as needed.

---

# API Documentation

Swagger/OpenAPI documentation is generated from protobuf definitions.

Example endpoints:

```text
POST   /v1/files/upload-url
POST   /v1/files/multipart
POST   /v1/files/multipart/url
POST   /v1/files/multipart/complete
POST   /v1/files/multipart/abort

GET    /v1/files/{file_id}
POST   /v1/files/{file_id}/download-url
POST   /v1/files/{file_id}/preview-url

DELETE /v1/files/{file_id}
```

---

# Upload Flow

## Simple Upload

```text
Client
  └── Request upload URL
        └── Upload directly to object storage
```

---

## Multipart Upload

```text
Create multipart upload
        ↓
Generate upload URL per chunk
        ↓
Upload chunks directly to object storage
        ↓
Complete multipart upload
```

---

# Docker Compose

Example local infrastructure:

```yaml
services:
  minio:
    image: minio/minio:latest

  mysql:
    image: mysql:8.4

  clamav:
    image: mkodockx/docker-clamav:alpine
```

---

# Future Roadmap

- Resumable upload
- Upload session recovery
- Async virus scanning worker
- Multi-tenant support
- Upload rate limiting
- Object lifecycle management
- Background cleanup worker
- File versioning
- Storage abstraction improvements

---

# Deployment

This service is designed to be deployed in cloud-native environments.

Recommended production stack:

- Kubernetes / GKE
- Managed MySQL / Cloud SQL
- S3-compatible object storage
- External secret management
- Horizontal scaling

---

# License

MIT License.