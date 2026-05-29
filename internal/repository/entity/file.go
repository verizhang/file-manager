package entity

import (
	"time"

	"github.com/verizhang/file-manager/internal/model"
	"gorm.io/gorm"
)

type File struct {
	ID              string                `gorm:"column:id;type:varchar(36);primaryKey"`
	UploadID        *string               `gorm:"column:upload_id;type:varchar(255)"`
	Bucket          string                `gorm:"column:bucket;type:varchar(255)"`
	ObjectKey       string                `gorm:"column:object_key;type:varchar(255);not null"`
	FileName        string                `gorm:"column:file_name;type:varchar(255);not null"`
	ContentType     string                `gorm:"column:content_type;type:varchar(100);not null"`
	Size            int64                 `gorm:"column:size;not null"`
	ETag            *string               `gorm:"column:etag;type:varchar(255)"`
	Status          model.FileStatus      `gorm:"column:status;type:varchar(50);not null"`
	VirusScanStatus model.VirusScanStatus `gorm:"column:virus_scan_status;type:varchar(50);not null"`
	CreatedAt       time.Time             `gorm:"column:created_at"`
	UpdatedAt       time.Time             `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt        `gorm:"column:deleted_at;index"`
}

func (File) TableName() string {
	return "files"
}
