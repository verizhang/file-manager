package model

type FileStatus string

const (
	FileStatusPending     FileStatus = "pending"
	FileStatusUploading   FileStatus = "uploading"
	FileStatusPendingScan FileStatus = "pending_scan"
	FileStatusCompleted   FileStatus = "completed"
	FileStatusFailed      FileStatus = "failed"
	FileStatusDeleted     FileStatus = "deleted"
	FileStatusInfected    FileStatus = "infected"
)
