# File Manager Service

Production-oriented file management microservice built with Go, gRPC, MySQL, and S3-compatible object storage.

This project demonstrates how modern backend services handle file uploads at scale using pre-signed URLs, multipart uploads, metadata persistence, and object storage abstraction.

---

## Overview

Instead of uploading files through the application server, clients upload files directly to object storage using pre-signed URLs.

The service is responsible for:

* File metadata management
* Upload lifecycle management
* Pre-signed upload URL generation
* Multipart upload orchestration
* Download URL generation
* Preview URL generation
* Object storage abstraction
* Upload status tracking

This architecture reduces application server load and scales significantly better for large files.

---

## Features

### Upload Management

* Direct upload using S3 pre-signed URLs
* Multipart upload support
* Upload completion workflow
* Upload abort workflow
* File lifecycle tracking

### File Access

* Generate temporary download URLs
* Generate temporary preview URLs
* Retrieve file metadata

### Storage

* S3-compatible storage abstraction
* MinIO support
* AWS S3 support

### Metadata

* File status tracking
* Virus scan status tracking
* Upload timestamps
* Soft deletion support

---

## Architecture

```text
                 ┌──────────────┐
                 │    Client    │
                 └──────┬───────┘
                        │
                        ▼
               HTTP / gRPC API
                        │
                        ▼
                 Handler Layer
                        │
                        ▼
                 Service Layer
                  Business Logic
                        │
            ┌───────────┴───────────┐
            ▼                       ▼
     Repository Layer       Storage Layer
            │                       │
            ▼                       ▼
         MySQL              S3 / MinIO
```

### Layer Responsibilities

#### Handler

Responsible for:

* Request validation
* gRPC implementation
* Response mapping
* Error translation

#### Service

Responsible for:

* Business rules
* Upload workflows
* Metadata processing
* File lifecycle management

#### Repository

Responsible for:

* Database access
* Persistence logic
* Query abstraction

#### Storage

Responsible for:

* S3 operations
* Pre-signed URLs
* Multipart uploads
* Object management

---

## Technology Stack

### Backend

* Go
* gRPC
* gRPC-Gateway
* Protocol Buffers
* GORM

### Database

* MySQL

### Object Storage

* AWS SDK v2
* Amazon S3
* MinIO

### Infrastructure

* Docker
* Docker Compose
* Buf

### Logging

* Zap Logger

---

## API Endpoints

### Simple Upload

```http
POST /v1/files/upload-url
```

Generate a pre-signed URL for direct upload.

---

### Complete Upload

```http
POST /v1/files/{file_id}/complete
```

Mark upload as completed and update metadata.

---

### Multipart Upload

Create upload session:

```http
POST /v1/files/multipart
```

Generate part upload URL:

```http
POST /v1/files/multipart/url
```

Complete multipart upload:

```http
POST /v1/files/multipart/complete
```

Abort multipart upload:

```http
POST /v1/files/multipart/abort
```

---

### File Access

Get file metadata:

```http
GET /v1/files/{file_id}
```

Generate download URL:

```http
POST /v1/files/{file_id}/download-url
```

Generate preview URL:

```http
POST /v1/files/{file_id}/preview-url
```

---

### Delete File

```http
DELETE /v1/files/{file_id}
```

---

## Project Structure

```text
cmd/
└── api/
    └── main.go

internal/
├── config/
├── database/
├── errs/
├── handler/
├── interceptor/
├── logger/
├── model/
├── repository/
├── server/
├── service/
└── storage/

proto/
└── file/v1/

gen/
├── go/
└── openapi/

migrations/
```

---

## Local Development

### Requirements

* Go 1.25+
* Docker
* Docker Compose
* MySQL
* MinIO

---

### Start Infrastructure

```bash
docker compose up -d
```

Services:

* MySQL
* MinIO

---

### Configure Environment

Create a `.env` file:

```env
APP_HTTP_PORT=8080
APP_GRPC_PORT=9090

DB_HOST=localhost
DB_PORT=3306
DB_NAME=file_manager
DB_USER=mysqladmin
DB_PASSWORD=mysqladmin

S3_ENDPOINT=http://localhost:9000
S3_REGION=ap-southeast-1
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=file-manager
S3_USE_SSL=false
```

---

### Run Migration

```bash
migrate up
```

---

### Generate Protobuf

```bash
buf generate
```

---

### Run Service

```bash
go run cmd/api/main.go
```

---

## Example Upload Flow

### 1. Request Upload URL

```http
POST /v1/files/upload-url
```

Response:

```json
{
  "file_id": "c1c57d4d-9b93-4c62-bf92-fd56a7f89f21",
  "upload_url": "https://...",
  "object_key": "uploads/2026/05/example.pdf"
}
```

### 2. Upload Directly to Storage

```bash
curl -X PUT \
  --upload-file example.pdf \
  "PRESIGNED_URL"
```

### 3. Complete Upload

```http
POST /v1/files/{file_id}/complete
```

### 4. Retrieve Metadata

```http
GET /v1/files/{file_id}
```

---

## Future Improvements

* Authentication & Authorization
* ClamAV Integration
* Background Processing
* Event-Driven Architecture
* OpenTelemetry Tracing
* Prometheus Metrics
* Kubernetes Deployment
* CI/CD Pipeline
* Object Lifecycle Policies

---

## Learning Goals

This project was built to explore:

* gRPC service development
* gRPC-Gateway integration
* Clean architecture principles
* Repository pattern
* S3 object storage workflows
* Multipart upload design
* Production-oriented backend development
* Service abstraction and dependency injection

---

## License

MIT
