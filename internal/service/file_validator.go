package service

import (
	"fmt"

	"github.com/verizhang/file-manager/internal/errs"
)

func (s *fileService) validateContentType(
	contentType string,
) error {

	for _, allowedType := range s.cfg.File.AllowedTypes {
		if allowedType == contentType {
			return nil
		}
	}

	return fmt.Errorf("%w: %s", errs.ErrFileTypeNotAllowed, fmt.Sprintf("content type %s not allowed", contentType))
}

func (s *fileService) validateFileSize(
	size int64,
) error {

	if size > s.cfg.File.MaxFileSize {
		return fmt.Errorf("%w: %s", errs.ErrFileTooLarge, fmt.Errorf("size %s too large"))
	}

	return nil
}
