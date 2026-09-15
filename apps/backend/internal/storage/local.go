package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type LocalStorage struct {
	basePath  string
	secretKey []byte
	serveRoot string
}

func NewLocalStorage(basePath, secretKey, serveRoot string) (*LocalStorage, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve storage path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	if secretKey == "" {
		secretKey = "default-dev-secret-change-me"
	}

	if serveRoot == "" {
		serveRoot = "/api/files/serve"
	}

	return &LocalStorage{
		basePath:  absPath,
		secretKey: []byte(secretKey),
		serveRoot: serveRoot,
	}, nil
}

func (s *LocalStorage) Put(ctx context.Context, key string, reader io.Reader, size int64) error {
	path := s.resolvePath(key)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	written, err := io.Copy(f, reader)
	if err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if size > 0 && written != size {
		os.Remove(tmpPath)
		return fmt.Errorf("size mismatch: expected %d bytes, wrote %d", size, written)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (s *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	path := s.resolvePath(key)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return f, nil
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	path := s.resolvePath(key)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *LocalStorage) URL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	exp := time.Now().Add(expiry).Unix()
	sig := s.sign(key, exp)

	return fmt.Sprintf("%s/%s?exp=%d&sig=%s", s.serveRoot, key, exp, sig), nil
}

func (s *LocalStorage) Stat(ctx context.Context, key string) (int64, error) {
	path := s.resolvePath(key)

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("file not found: %s", key)
		}
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}

	return info.Size(), nil
}

func (s *LocalStorage) VerifyURL(key string, exp int64, sig string) bool {
	if time.Now().Unix() > exp {
		return false
	}
	expected := s.sign(key, exp)
	return hmac.Equal([]byte(sig), []byte(expected))
}

func (s *LocalStorage) resolvePath(key string) string {
	clean := filepath.Clean(key)
	clean = strings.TrimPrefix(clean, "/")
	return filepath.Join(s.basePath, clean)
}

func (s *LocalStorage) sign(key string, exp int64) string {
	mac := hmac.New(sha256.New, s.secretKey)
	mac.Write([]byte(key))
	mac.Write([]byte(strconv.FormatInt(exp, 10)))
	return hex.EncodeToString(mac.Sum(nil))
}

func ParseSignedURL(rawURL string) (key string, exp int64, sig string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid URL: %w", err)
	}

	key = strings.TrimPrefix(u.Path, "/")
	expStr := u.Query().Get("exp")
	sig = u.Query().Get("sig")

	if expStr == "" || sig == "" {
		return "", 0, "", fmt.Errorf("missing exp or sig parameter")
	}

	exp, err = strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid exp parameter: %w", err)
	}

	return key, exp, sig, nil
}
