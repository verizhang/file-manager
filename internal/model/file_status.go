package model

type FileStatus string

const (
	FileStatusPending   FileStatus = "PENDING"
	FileStatusUploading FileStatus = "UPLOADING"
	FileStatusCompleted FileStatus = "COMPLETED"
	FileStatusAborted   FileStatus = "ABORTED"
)
