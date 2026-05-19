# File Manager Service

A production-oriented file management microservice built with Golang, gRPC, gRPC-Gateway, MySQL, and S3-compatible object storage such as MinIO or AWS S3.

This project supports:

* presigned upload URL generation
* multipart upload flow
* preview & download URL generation
* resumable downloads
* file metadata management
* virus scanning preparation
* scalable storage abstraction

---

# Features

## Upload

* Simple upload using presigned URL
* Multipart upload support
* S3-compatible storage
* Object key generation strategy

## File Access

* Generate preview URL
* Generate download URL
* Get file metadata

## File Management

* Delete file
* File status tracking
* Upload lifecycle tracking

## Architecture

* Clean Architecture-ish layering
* Repository pattern
* Storage abstraction
* gRPC + HTTP Gateway
* Migration-based database management

---

# Tech Stack

## Backend

* Golang
* gRPC
* gRPC-Gateway
* GORM
* MySQL

## Storage

* AWS SDK v2
* MinIO
* AWS S3 compatible storage

## Infrastructure

* Docker
* Docker Compose

## Planned

* Kubernetes
* CI/CD
* Virus scanning (ClamAV)
* Multipart resumable upload optimization

---

# Project Structure

```text
.
├── cmd/
│   └── api/
│       └── main.go
│
├── gen/
│   ├── go/
│   ├── grpc-gateway/
│   └── openapiv2/
│
├── internal/
│   ├── config/
│   ├── database/
│   ├── handler/
│   │   └── grpc/
│   ├── logger/
│   ├── model/
│   ├── repository/
│   │   ├── entity/
│   │   ├── mapper/
│   │   └── mysql/
│   ├── server/
│   ├── service/
│   └── storage/
│       └── s3/
│
├── migrations/
│
├── proto/
│
├── scripts/
│
├── docker-compose.yml
├── Makefile
├── buf.gen.yaml
├── buf.yaml
└── README.md
```

---

# Architecture

```text
Client
   ↓
HTTP / gRPC
   ↓
Handler Layer
   ↓
Service Layer
   ↓
Repository Layer
   ↓
MySQL

Service Layer
   ↓
Storage Abstraction
   ↓
S3 Compatible Storage
(MinIO / AWS S3)
```

---

# API Overview

## Simple Upload

* `POST /v1/files/upload-url`

Generate presigned upload URL for direct upload.

---

## Multipart Upload

* `POST /v1/files/multipart`
* `POST /v1/files/multipart/url`
* `POST /v1/files/multipart/complete`
* `POST /v1/files/multipart/abort`

Multipart upload lifecycle.

---

## File Access

* `GET /v1/files/{file_id}`
* `POST /v1/files/{file_id}/download-url`
* `POST /v1/files/{file_id}/preview-url`

---

## Delete

* `DELETE /v1/files/{file_id}`

---

# Local Development

## Requirements

* Go 1.24+
* Docker
* Docker Compose
* Make
* golang-migrate

---

# Run Infrastructure

```bash
docker compose up -d
```

This will start:

* MySQL
* MinIO

---

# Create Bucket

Open MinIO console:

```text
http://localhost:9001
```

Default credentials:

```text
username: minioadmin
password: minioadmin
```

Create bucket:

```text
file-manager
```

---

# Environment Variables

Create `.env`

```env
# =====================================================
# APP
# =====================================================

APP_NAME=file-manager
APP_ENV=local

APP_HTTP_PORT=8080
APP_GRPC_PORT=9090

# =====================================================
# DATABASE
# =====================================================

DB_HOST=localhost
DB_PORT=3306
DB_NAME=file_manager

DB_USER=mysqladmin
DB_PASSWORD=mysqladmin

DB_MAX_OPEN_CONNECTION=10
DB_MAX_IDLE_CONNECTION=5
DB_MAX_LIFETIME_MINUTE=30

# =====================================================
# S3 / MINIO
# =====================================================

S3_ENDPOINT=http://localhost:9000
S3_REGION=ap-southeast-1

S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin

S3_BUCKET=file-manager

S3_USE_SSL=false

# =====================================================
# CLAMAV
# =====================================================

CLAMAV_HOST=localhost
CLAMAV_PORT=3310
```

---

# Run Database Migration

```bash
migrate \
  -path migrations \
  -database "mysql://mysqladmin:mysqladmin@tcp(localhost:3306)/file_manager" \
  up
```

---

# Generate Protobuf

```bash
buf generate
```

---

# Run Application

```bash
go run cmd/api/main.go
```

---

# Test Upload URL Endpoint

```bash
curl --location 'http://localhost:8080/v1/files/upload-url' \
--header 'Content-Type: application/json' \
--data '{
  "file_name": "example.pdf",
  "content_type": "application/pdf",
  "size": 1048576
}'
```

---

# Example Response

```json
{
  "file_id": "uuid",
  "upload_url": "https://...",
  "object_key": "2026/05/19/uuid.pdf"
}
```

---

# Upload File Using Presigned URL

```bash
curl --request PUT \
--upload-file ./example.pdf \
--header "Content-Type: application/pdf" \
'PRESIGNED_URL'
```

---

# Current Status

## Implemented

* Project structure
* gRPC server
* gRPC Gateway
* S3 storage abstraction
* Presigned upload URL generation
* MySQL repository layer
* Migration system
* File metadata persistence

## In Progress

* Multipart upload implementation
* Download URL generation
* Preview URL generation
* Delete flow

## Planned

* Virus scanning
* Authentication middleware
* Background workers
* Event-driven processing
* Metrics & observability
* Kubernetes deployment
* CI/CD pipeline
* Unit testing

---

# Design Goals

This project is designed to:

* be cloud-native ready
* support large-scale file uploads
* support resumable upload flows
* work with any S3-compatible provider
* be easy to deploy for companies
* provide production-oriented architecture patterns

---

# License

MIT License
