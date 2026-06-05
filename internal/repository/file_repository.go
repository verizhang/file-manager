package repository

import (
	"context"

	"github.com/verizhang/file-manager/internal/model"
)

type FileRepository interface {
	Create(
		ctx context.Context,
		file *model.File,
	) error
	GetByID(
		ctx context.Context,
		id string,
	) (*model.File, error)

	UpdateStatus(
		ctx context.Context,
		id string,
		status model.FileStatus,
	) error
	GetByObjectKey(
		ctx context.Context,
		objectKey string,
	) (*model.File, error)
	UpdateStatusAndETag(
		ctx context.Context,
		id string,
		status model.FileStatus,
		etag string,
	) error
	UpdateStatusAndClearUploadID(
		ctx context.Context,
		id string,
		status model.FileStatus,
	) error
	Delete(
		ctx context.Context,
		id string,
	) error
}
