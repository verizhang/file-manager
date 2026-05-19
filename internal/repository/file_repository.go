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
}
