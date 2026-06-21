package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Endpoint  string `envconfig:"S3_ENDPOINT" required:"true"`
	AccessKey string `envconfig:"S3_ACCESS_KEY" required:"true"`
	SecretKey string `envconfig:"S3_SECRET_KEY" required:"true"`
	Bucket    string `envconfig:"S3_BUCKET" required:"true"`
	UseSSL    bool   `envconfig:"S3_USE_SSL" default:"false"`
	Region    string `envconfig:"S3_REGION" default:"us-east-1"`
}

func NewClient(
	ctx context.Context,
	cfg Config,
) (*awss3.Client, error) {

	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				HostnameImmutable: true,
			}, nil
		},
	)

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
		),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithResponseChecksumValidation(aws.ResponseChecksumValidationWhenRequired),
	)
	if err != nil {
		return nil, err
	}

	return awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		o.UsePathStyle = true
	}), nil
}
