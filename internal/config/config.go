package config

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	database "github.com/verizhang/file-manager/pkg/database/mysql"
	"github.com/verizhang/file-manager/pkg/messaging/rabbitmq"
	"github.com/verizhang/file-manager/pkg/storage/s3"
	"github.com/verizhang/file-manager/pkg/virusscanner/clamav"
)

type Config struct {
	App             AppConfig
	DB              database.MySQLConfig
	S3              s3.Config
	PresignedConfig PresignedConfig
	Multipart       MultipartConfig
	File            FileConfig
	ClamAV          clamav.Config
	RateLimit       RateLimitConfig
	RabbitMQ        rabbitmq.Config
}

type AppConfig struct {
	Name     string `envconfig:"APP_NAME" default:"file-manager"`
	Env      string `envconfig:"APP_ENV" default:"development"`
	HTTPPort int    `envconfig:"APP_HTTP_PORT" default:"8080"`
	GRPCPort int    `envconfig:"APP_GRPC_PORT" default:"9090"`
	Debug    bool   `envconfig:"APP_DEBUG" default:"true"`
}

type PresignedConfig struct {
	UploadExpireMinutes   time.Duration `envconfig:"PRESIGNED_UPLOAD_EXPIRE_MINUTES" default:"15"`
	DownloadExpireMinutes time.Duration `envconfig:"PRESIGNED_DOWNLOAD_EXPIRE_MINUTES" default:"30"`
}

type MultipartConfig struct {
	PartSize int64 `envconfig:"MULTIPART_PART_SIZE" default:"5242880"`
}

type FileConfig struct {
	MaxFileSize  int64    `envconfig:"MAX_FILE_SIZE" default:"104857600"`
	AllowedTypes []string `envconfig:"ALLOWED_FILE_TYPES"`
}

type RateLimitConfig struct {
	Request  int           `envconfig:"RATE_LIMIT_REQUEST" default:"100"`
	Duration time.Duration `envconfig:"RATE_LIMIT_DURATION" default:"1m"`
}

func Load() *Config {
	cfg := &Config{}

	err := envconfig.Process("", &cfg.App)
	if err != nil {
		log.Fatalf("failed load app config: %v", err)
	}

	err = envconfig.Process("", &cfg.DB)
	if err != nil {
		log.Fatalf("failed load db config: %v", err)
	}

	err = envconfig.Process("", &cfg.S3)
	if err != nil {
		log.Fatalf("failed load s3 config: %v", err)
	}

	err = envconfig.Process("", &cfg.PresignedConfig)
	if err != nil {
		log.Fatalf("failed load presigned config: %v", err)
	}

	err = envconfig.Process("", &cfg.Multipart)
	if err != nil {
		log.Fatalf("failed load multipart config: %v", err)
	}

	err = envconfig.Process("", &cfg.File)
	if err != nil {
		log.Fatalf("failed load file config: %v", err)
	}

	err = envconfig.Process("", &cfg.ClamAV)
	if err != nil {
		log.Fatalf("failed load clamav config: %v", err)
	}

	err = envconfig.Process("", &cfg.RateLimit)
	if err != nil {
		log.Fatalf("failed load rate limit config: %v", err)
	}

	err = envconfig.Process("", &cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("failed load RabbitMQ config: %v", err)
	}

	return cfg
}
