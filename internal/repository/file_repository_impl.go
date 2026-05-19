package repository

import (
	"context"
	"gorm.io/gorm"
	"github.com/verizhang/file-manager/internal/model"
)

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(
	ctx context.Context,
	file *model.File,
) error {

	// TODO: implement database insert

	return nil
}

func (r *fileRepository) FindByID(
	ctx context.Context,
	id string,
) (*model.File, error) {

	// TODO: implement database query

	return nil, nil
}