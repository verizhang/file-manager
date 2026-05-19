package mysql

import (
	"context"

	"github.com/verizhang/file-manager/internal/model"
	"github.com/verizhang/file-manager/internal/repository"
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
		return err
	}

	return nil
}
