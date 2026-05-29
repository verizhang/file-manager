package mapper

import (
	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
	"github.com/verizhang/file-manager/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoFile(
	file *model.File,
) *filev1.FileObject {

	var etag string

	if file.ETag != nil {
		etag = *file.ETag
	}

	return &filev1.FileObject{
		Id:              file.ID,
		Bucket:          file.Bucket,
		ObjectKey:       file.ObjectKey,
		FileName:        file.FileName,
		ContentType:     file.ContentType,
		Size:            file.Size,
		Etag:            etag,
		Status:          toProtoFileStatus(file.Status),
		VirusScanStatus: toProtoVirusScanStatus(file.VirusScanStatus),
		CreatedAt:       timestamppb.New(file.CreatedAt),
	}
}

func toProtoFileStatus(
	status model.FileStatus,
) filev1.FileStatus {
	switch status {
	case model.FileStatusPending:
		return filev1.FileStatus_FILE_STATUS_PENDING

	case model.FileStatusUploading:
		return filev1.FileStatus_FILE_STATUS_UPLOADING

	case model.FileStatusCompleted:
		return filev1.FileStatus_FILE_STATUS_COMPLETED

	case model.FileStatusAborted:
		return filev1.FileStatus_FILE_STATUS_ABORTED

	default:
		return filev1.FileStatus_FILE_STATUS_UNSPECIFIED
	}
}

func toProtoVirusScanStatus(
	status model.VirusScanStatus,
) filev1.VirusScanStatus {
	switch status {
	case model.VirusScanStatusPending:
		return filev1.VirusScanStatus_VIRUS_SCAN_STATUS_PENDING
	case model.VirusScanStatusScaning:
		return filev1.VirusScanStatus_VIRUS_SCAN_STATUS_SCANNING
	case model.VirusScanStatusClean:
		return filev1.VirusScanStatus_VIRUS_SCAN_STATUS_CLEAN
	case model.VirusScanStatusInfected:
		return filev1.VirusScanStatus_VIRUS_SCAN_STATUS_INFECTED
	case model.VirusScanStatusFailed:
		return filev1.VirusScanStatus_VIRUS_SCAN_STATUS_FAILED
	default:
		return filev1.VirusScanStatus_VIRUS_SCAN_STATUS_UNSPECIFIED
	}
}
