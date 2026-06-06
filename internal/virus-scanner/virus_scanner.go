package virusscanner

import (
	"context"
	"io"
)

type Status string

const (
	StatusClean    Status = "CLEAN"
	StatusInfected Status = "INFECTED"
)

type ScanOptions struct {
	FileName string
	Reader   io.Reader
}

type ScanResult struct {
	Status    Status
	Threat    string
	RawResult string
}

type Scanner interface {
	Scan(
		ctx context.Context,
		opts ScanOptions,
	) (*ScanResult, error)
}
