package model

import "time"

type File struct {
	ID          string
	UploadID    *string
	Bucket      string
	ObjectKey   string
	FileName    string
	ContentType string
	Size        int64
	ETag        *string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
