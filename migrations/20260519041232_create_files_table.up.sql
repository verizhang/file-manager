CREATE TABLE files (
    id VARCHAR(36) PRIMARY KEY,

    upload_id VARCHAR(255) NULL,

    object_key VARCHAR(255) NOT NULL,

    bucket VARCHAR(255) NOT NULL,

    file_name VARCHAR(255) NOT NULL,

    content_type VARCHAR(100) NOT NULL,

    size BIGINT NOT NULL,

    etag VARCHAR(255) NULL,

    status VARCHAR(50) NOT NULL,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_files_deleted_at
ON files(deleted_at);

CREATE INDEX idx_files_status
ON files(status);

CREATE INDEX idx_files_upload_id
ON files(upload_id);