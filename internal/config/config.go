package config

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App       AppConfig
	DB        DBConfig
	S3        S3Config
	Presigned PresignedConfig
	Multipart MultipartConfig
	File      FileConfig
	ClamAV    ClamAVConfig
	RateLimit RateLimitConfig
}

type AppConfig struct {
	Name     string `envconfig:"APP_NAME" default:"file-manager"`
	Env      string `envconfig:"APP_ENV" default:"development"`
	HTTPPort int    `envconfig:"APP_HTTP_PORT" default:"8080"`
	GRPCPort int    `envconfig:"APP_GRPC_PORT" default:"9090"`
	Debug    bool   `envconfig:"APP_DEBUG" default:"true"`
}

type DBConfig struct {
	Host     string `envconfig:"DB_HOST" required:"true"`
	Port     int    `envconfig:"DB_PORT" default:"3306"`
	User     string `envconfig:"DB_USER" required:"true"`
	Password string `envconfig:"DB_PASSWORD"`
	Name     string `envconfig:"DB_NAME" required:"true"`
	TLS      string `envconfig:"DB_TLS" default:"skip-verify"`
}

type S3Config struct {
	Endpoint        string `envconfig:"S3_ENDPOINT" required:"true"`
	AccessKey       string `envconfig:"S3_ACCESS_KEY" required:"true"`
	SecretKey       string `envconfig:"S3_SECRET_KEY" required:"true"`
	Bucket          string `envconfig:"S3_BUCKET" required:"true"`
	UseSSL          bool   `envconfig:"S3_USE_SSL" default:"false"`
	Region          string `envconfig:"S3_REGION" default:"us-east-1"`
	PresignedConfig PresignedConfig
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

type ClamAVConfig struct {
	Enabled bool   `envconfig:"CLAMAV_ENABLED" default:"false"`
	Host    string `envconfig:"CLAMAV_HOST" default:"localhost"`
	Port    int    `envconfig:"CLAMAV_PORT" default:"3310"`
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

	err = envconfig.Process("", &cfg.Presigned)
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

	return cfg
}
