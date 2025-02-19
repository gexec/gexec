package upload

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	transport "github.com/aws/smithy-go/endpoints"
	"github.com/gexec/gexec/pkg/config"
)

// S3Upload implements the Upload interface.
type S3Upload struct {
	endpoint string
	path     string
	access   string
	secret   string
	bucket   string
	region   string

	client  *s3.Client
	presign *s3.PresignClient
}

// Info prepares some informational message about the handler.
func (u *S3Upload) Info() map[string]interface{} {
	result := make(map[string]interface{})
	result["driver"] = "s3"
	result["endpoint"] = u.endpoint
	result["path"] = u.path
	result["bucket"] = u.bucket
	result["region"] = u.region

	return result
}

// Prepare simply prepares the upload handler.
func (u *S3Upload) Prepare() (Upload, error) {
	globalOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(
			u.region,
		),
	}

	if u.access != "" && u.secret != "" {
		access, err := config.Value(u.access)

		if err != nil {
			return nil, fmt.Errorf("failed to parse access key: %w", err)
		}

		secret, err := config.Value(u.secret)

		if err != nil {
			return nil, fmt.Errorf("failed to parse secret key: %w", err)
		}

		globalOpts = append(
			globalOpts,
			awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					access,
					secret,
					"",
				),
			),
		)
	}

	cfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		globalOpts...,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	serviceOpts := []func(*s3.Options){}

	if u.endpoint != "" {
		endpoint, err := url.Parse(u.endpoint)

		if err != nil {
			return nil, err
		}

		serviceOpts = append(
			serviceOpts,
			s3.WithEndpointResolverV2(
				&CustomEndpointResolver{
					Endpoint: endpoint,
				},
			),
		)
	}

	u.client = s3.NewFromConfig(
		cfg,
		serviceOpts...,
	)

	u.presign = s3.NewPresignClient(
		u.client,
	)

	return u, nil
}

// Close simply closes the upload handler.
func (u *S3Upload) Close() error {
	return nil
}

// Upload stores an attachment within the defined S3 bucket.
func (u *S3Upload) Upload(ctx context.Context, key, ctype string, content []byte) error {
	params := &s3.PutObjectInput{
		ACL:         types.ObjectCannedACLPublicRead,
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(path.Join(u.path, key)),
		ContentType: aws.String(ctype),
		Body:        bytes.NewReader(content),
	}

	if _, err := u.client.PutObject(
		ctx,
		params,
	); err != nil {
		return err
	}

	return nil
}

// Delete removes an attachment from the defined S3 bucket.
func (u *S3Upload) Delete(ctx context.Context, key string) error {
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(path.Join(u.path, key)),
	}

	if _, err := u.client.DeleteObject(
		ctx,
		params,
	); err != nil {
		return err
	}

	return nil
}

// Handler implements an HTTP handler for asset uploads.
func (u *S3Upload) Handler(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := u.presign.PresignGetObject(
			r.Context(),
			&s3.GetObjectInput{
				Bucket: aws.String(u.bucket),
				Key:    aws.String(path.Join(u.path, strings.TrimPrefix(r.URL.Path, root))),
			},
			func(opts *s3.PresignOptions) {
				opts.Expires = time.Duration(5 * time.Minute)
			},
		)

		if err != nil {
			http.Error(
				w,
				http.StatusText(http.StatusNotFound),
				http.StatusNotFound,
			)
		}

		http.Redirect(w, r, req.URL, http.StatusTemporaryRedirect)
	})
}

// NewS3Upload initializes a new S3 handler.
func NewS3Upload(cfg config.Upload) (Upload, error) {
	path := "/"

	if cfg.Path != "" {
		path = cfg.Path
	}

	f := &S3Upload{
		endpoint: cfg.Endpoint,
		path:     path,
		access:   cfg.Access,
		secret:   cfg.Secret,
		bucket:   cfg.Bucket,
		region:   cfg.Region,
	}

	return f.Prepare()
}

// MustS3Upload simply calls NewS3Upload and panics on an error.
func MustS3Upload(cfg config.Upload) Upload {
	db, err := NewS3Upload(cfg)

	if err != nil {
		panic(err)
	}

	return db
}

// CustomEndpointResolver is used for S3 compatible storage endpoints.
type CustomEndpointResolver struct {
	Endpoint *url.URL
}

// ResolveEndpoint resolves endpoints for a specific service and region
func (r *CustomEndpointResolver) ResolveEndpoint(_ context.Context, _ s3.EndpointParameters) (transport.Endpoint, error) {
	return transport.Endpoint{
		URI: *r.Endpoint,
	}, nil
}
