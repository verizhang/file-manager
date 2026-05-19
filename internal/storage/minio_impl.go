package storage

import "context"

func (m *Minio) CreateMultipartUpload(
	ctx context.Context,
	objectKey string,
	contentType string,
) (string, error) {

	// TODO:
	// call CreateMultipartUpload from MinIO SDK

	return "", nil
}