package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// R2Storage provides interface for Cloudflare R2 operations
type R2Storage interface {
	// GeneratePresignedUploadURL generates a presigned URL for direct upload
	GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiresIn time.Duration) (string, error)

	// UploadFile uploads a file directly from backend (fallback method)
	UploadFile(ctx context.Context, file io.Reader, key string, contentType string) (string, error)

	// DeleteFile deletes a file from R2
	DeleteFile(ctx context.Context, key string) error

	// GetPublicURL returns the public CDN URL for a file
	GetPublicURL(key string) string

	// GeneratePresignedDownloadURL generates a presigned URL for private file access
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiresIn time.Duration) (string, error)
}

type R2StorageImpl struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucketName string
	publicURL  string // e.g., https://pub-xxxxx.r2.dev
}

// NewR2Storage creates a new R2 storage client
func NewR2Storage(accountID, accessKeyID, secretAccessKey, bucketName, publicURL string) (R2Storage, error) {
	// Validate inputs
	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return nil, fmt.Errorf("missing required R2 configuration")
	}

	// Create R2 endpoint URL
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	// Create AWS config with R2 endpoint
	cfg := aws.Config{
		Region: "auto", // R2 uses "auto" region
		Credentials: credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		),
	}

	// Create S3 client with R2 endpoint
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	// Create presigner
	presigner := s3.NewPresignClient(client)

	return &R2StorageImpl{
		client:     client,
		presigner:  presigner,
		bucketName: bucketName,
		publicURL:  publicURL,
	}, nil
}

// GeneratePresignedUploadURL generates a presigned URL for PUT upload
func (r *R2StorageImpl) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiresIn time.Duration) (string, error) {
	// Create PutObject input
	input := &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	// Generate presigned PUT URL
	presignedReq, err := r.presigner.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return presignedReq.URL, nil
}

// UploadFile uploads a file directly from backend
func (r *R2StorageImpl) UploadFile(ctx context.Context, file io.Reader, key string, contentType string) (string, error) {
	// Upload to R2
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead, // Make publicly readable
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to R2: %w", err)
	}

	// Return public URL
	return r.GetPublicURL(key), nil
}

// DeleteFile deletes a file from R2
func (r *R2StorageImpl) DeleteFile(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}
	return nil
}

// GetPublicURL returns the public CDN URL for a file
func (r *R2StorageImpl) GetPublicURL(key string) string {
	if r.publicURL == "" {
		// Fallback to R2.dev domain if no custom domain
		return fmt.Sprintf("https://%s.r2.dev/%s", r.bucketName, key)
	}
	return fmt.Sprintf("%s/%s", r.publicURL, key)
}

// GeneratePresignedDownloadURL generates a presigned URL for private file access
func (r *R2StorageImpl) GeneratePresignedDownloadURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	// Create GetObject input
	input := &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	// Generate presigned GET URL
	presignedReq, err := r.presigner.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return presignedReq.URL, nil
}

// Verify interface implementation
var _ R2Storage = (*R2StorageImpl)(nil)
