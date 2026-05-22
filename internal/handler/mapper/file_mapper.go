package mapper

import (
	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
	"github.com/verizhang/file-manager/internal/model"
)

func ToProtoFile(
	file *model.File,
) *filev1.FileObject {

	var etag string

	if file.ETag != nil {
		etag = *file.ETag
	}

	return &filev1.FileObject{
		Id:          file.ID,
		Bucket:      file.Bucket,
		ObjectKey:   file.ObjectKey,
		FileName:    file.FileName,
		ContentType: file.ContentType,
		Size:        file.Size,
		Etag:        etag,
		Status:      ToProtoFileStatus(file.Status),
	}
}

func ToProtoFileStatus(
	status model.FileStatus,
) filev1.FileStatus {

	switch status {

	case model.FileStatusPending:
		return filev1.FileStatus_FILE_STATUS_PENDING

	case model.FileStatusUploading:
		return filev1.FileStatus_FILE_STATUS_UPLOADING

	case model.FileStatusPendingScan:
		return filev1.FileStatus_FILE_STATUS_PENDING_SCAN

	case model.FileStatusCompleted:
		return filev1.FileStatus_FILE_STATUS_COMPLETED

	case model.FileStatusFailed:
		return filev1.FileStatus_FILE_STATUS_FAILED

	case model.FileStatusDeleted:
		return filev1.FileStatus_FILE_STATUS_DELETED

	case model.FileStatusInfected:
		return filev1.FileStatus_FILE_STATUS_INFECTED

	default:
		return filev1.FileStatus_FILE_STATUS_UNSPECIFIED
	}
}
