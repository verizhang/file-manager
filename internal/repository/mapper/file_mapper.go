package mapper

import (
	"github.com/verizhang/file-manager/internal/model"
	"github.com/verizhang/file-manager/internal/repository/entity"
)

func ToFileModel(
	fileEntity *entity.File,
) *model.File {
	return &model.File{
		ID:              fileEntity.ID,
		UploadID:        fileEntity.UploadID,
		Bucket:          fileEntity.Bucket,
		ObjectKey:       fileEntity.ObjectKey,
		FileName:        fileEntity.FileName,
		ContentType:     fileEntity.ContentType,
		Size:            fileEntity.Size,
		ETag:            fileEntity.ETag,
		Status:          fileEntity.Status,
		VirusScanStatus: fileEntity.VirusScanStatus,
	}
}

func ToFileEntity(
	fileModel *model.File,
) *entity.File {

	fileEntity := &entity.File{
		ID:              fileModel.ID,
		UploadID:        fileModel.UploadID,
		Bucket:          fileModel.Bucket,
		ObjectKey:       fileModel.ObjectKey,
		FileName:        fileModel.FileName,
		ContentType:     fileModel.ContentType,
		Size:            fileModel.Size,
		ETag:            fileModel.ETag,
		Status:          fileModel.Status,
		VirusScanStatus: fileModel.VirusScanStatus,
	}
	return fileEntity
}
