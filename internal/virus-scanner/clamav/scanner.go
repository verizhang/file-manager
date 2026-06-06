package clamav

import (
	"context"
	"fmt"

	"github.com/dutchcoders/go-clamd"

	"github.com/verizhang/file-manager/internal/config"
	"github.com/verizhang/file-manager/internal/errs"
	virusscanner "github.com/verizhang/file-manager/internal/virus-scanner"
)

type Scanner struct {
	client *clamd.Clamd
}

func NewScanner(cfg config.ClamAVConfig) *Scanner {
	address := fmt.Sprintf(
		"%s://%s:%d",
		cfg.Network,
		cfg.Host,
		cfg.Port,
	)

	return &Scanner{
		client: clamd.NewClamd(address),
	}
}

func (s *Scanner) Scan(
	ctx context.Context,
	opts virusscanner.ScanOptions,
) (*virusscanner.ScanResult, error) {
	if opts.Reader == nil {
		return nil, fmt.Errorf("%w: clamav scan reader is nil", errs.ErrVirusScan)
	}

	done := make(chan bool)

	response, err := s.client.ScanStream(
		opts.Reader,
		done,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrVirusScan, err)
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %v", errs.ErrVirusScan, ctx.Err())

	case result := <-response:
		if result == nil {
			return nil, fmt.Errorf("%w: empty clamav response", errs.ErrVirusScan)
		}

		if result.Status == clamd.RES_FOUND {
			return &virusscanner.ScanResult{
				Status: virusscanner.StatusInfected,
				Threat: result.Description,
			}, nil
		}

		return &virusscanner.ScanResult{
			Status: virusscanner.StatusClean,
		}, nil
	}
}
