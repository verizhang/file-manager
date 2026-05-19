package model

import "time"

type File struct {
	ID          string
	ObjectKey   string
	FileName    string
	ContentType string
	Size        int64
	Status      string

	CreatedAt time.Time
	UpdatedAt time.Time
}