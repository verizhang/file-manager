package model

type File struct {
	ID              string
	UploadID        *string
	Bucket          string
	ObjectKey       string
	FileName        string
	ContentType     string
	Size            int64
	ETag            *string
	Status          FileStatus
	VirusScanStatus VirusScanStatus
}
