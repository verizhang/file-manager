package clamav

import (
	"context"
	"fmt"
	"time"

	"github.com/dutchcoders/go-clamd"

	"github.com/verizhang/file-manager/pkg/errs"
	virusscanner "github.com/verizhang/file-manager/pkg/virusscanner"
)

type Config struct {
	Enabled   bool          `envconfig:"CLAMAV_ENABLED" default:"false"`
	Host      string        `envconfig:"CLAMAV_HOST" default:"localhost"`
	Port      int           `envconfig:"CLAMAV_PORT" default:"3310"`
	Network   string        `envconfig:"CLAMAV_NETWORK" default:"tcp"`
	ChunkSize int           `envconfig:"CLAMAV_CHUNK_SIZE" default:"1048576"`
	Timeout   time.Duration `envconfig:"CLAMAV_TIMEOUT" default:"5m"`
}

type Scanner struct {
	client *clamd.Clamd
}

func NewScanner(cfg Config) *Scanner {
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
