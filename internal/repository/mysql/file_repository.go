package mysql

import (
	"context"
	"errors"
	"fmt"

	"github.com/verizhang/file-manager/internal/errs"
	"github.com/verizhang/file-manager/internal/model"
	"github.com/verizhang/file-manager/internal/repository"
	"github.com/verizhang/file-manager/internal/repository/entity"
	"github.com/verizhang/file-manager/internal/repository/mapper"

	"gorm.io/gorm"
)

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(
	db *gorm.DB,
) repository.FileRepository {
	return &fileRepository{
		db: db,
	}
}

func (r *fileRepository) Create(
	ctx context.Context,
	file *model.File,
) error {

	fileEntity := mapper.ToFileEntity(file)

	err := r.db.WithContext(ctx).
		Create(fileEntity).
		Error
	if err != nil {
		return fmt.Errorf("%w: %v", errs.ErrCreateFileMetadata, err)
	}

	return nil
}

func (r *fileRepository) GetByID(
	ctx context.Context,
	id string,
) (*model.File, error) {

	var fileEntity entity.File

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&fileEntity).
		Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrFileNotFound
		}

		return nil, err
	}

	return mapper.ToFileModel(&fileEntity), nil
}

func (r *fileRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status model.FileStatus,
) error {

	err := r.db.WithContext(ctx).
		Model(&entity.File{}).
		Where("id = ?", id).
		Update("status", status).
		Error

	if err != nil {
		return err
	}

	return nil
}
