package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	client *minio.Client
	bucket string
}

func NewS3Storage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3Storage, error) {
	creds := credentials.NewStaticV4(accessKey, secretKey, "")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  creds,
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &S3Storage{client: client, bucket: bucket}, nil
}

func (s *S3Storage) Put(ctx context.Context, key string, reader io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}
	return nil
}

func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	buf := make([]byte, 1)
	n, err := obj.Read(buf)
	if err != nil && err != io.EOF {
		obj.Close()
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	if n > 0 {
		obj.Close()
		obj2, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to re-open object: %w", err)
		}
		return &prependReader{buf[:n], obj2}, nil
	}

	return obj, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

func (s *S3Storage) URL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *S3Storage) Stat(ctx context.Context, key string) (int64, error) {
	info, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to stat object: %w", err)
	}
	return info.Size, nil
}

type prependReader struct {
	prefix []byte
	reader io.ReadCloser
}

func (r *prependReader) Read(p []byte) (int, error) {
	if len(r.prefix) > 0 {
		n := copy(p, r.prefix)
		r.prefix = r.prefix[n:]
		return n, nil
	}
	return r.reader.Read(p)
}

func (r *prependReader) Close() error {
	return r.reader.Close()
}
