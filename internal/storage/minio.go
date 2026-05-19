package storage

import (
	"github.com/minio/minio-go/v7"
)

type Minio struct {
	Client *minio.Client
	Bucket string
}

func New(
	client *minio.Client,
	bucket string,
) *Minio {
	return &Minio{
		Client: client,
		Bucket: bucket,
	}
}