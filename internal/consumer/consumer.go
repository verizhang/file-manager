package consumer

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/verizhang/file-manager/internal/model"
	"github.com/verizhang/file-manager/internal/service"
)

type Handler struct {
	logger      *zap.Logger
	fileService service.FileService
}

func NewHandler(
	logger *zap.Logger,
	fileService service.FileService,
) *Handler {
	return &Handler{
		logger:      logger,
		fileService: fileService,
	}
}

func (h *Handler) VirusScanner(
	ctx context.Context,
	payload []byte,
) error {

	var file model.File

	if err := json.Unmarshal(payload, &file); err != nil {
		h.logger.Error(
			"failed to unmarshal virus scan message",
			zap.Error(err),
		)

		return err
	}

	if file.ID == "" {
		h.logger.Warn(
			"virus scan message missing file id",
		)

		return nil
	}

	h.logger.Info(
		"processing virus scan request",
		zap.String("file_id", file.ID),
	)

	if err := h.fileService.ScanFile(
		ctx,
		file,
	); err != nil {
		h.logger.Error(
			"failed to scan file",
			zap.String("file_id", file.ID),
			zap.Error(err),
		)

		return err
	}

	return nil
}
