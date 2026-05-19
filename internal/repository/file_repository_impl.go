package repository

import (
	"context"

	"github.com/verizhang/file-manager/internal/model"
)

type fileRepository struct {
}

func NewFileRepository() FileRepository {
	return &fileRepository{}
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