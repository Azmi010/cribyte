// Package thumbnail generates small preview images for files.
package thumbnail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"os/exec"
	"time"

	"github.com/disintegration/imaging"

	_ "golang.org/x/image/webp"
)

const (
	MaxDimension     = 320
	jpegQuality      = 80
	videoSeekSeconds = 1
	externalTimeout  = 20 * time.Second
)

var ErrToolUnavailable = errors.New("thumbnail tool unavailable")

type Options struct {
	FfmpegPath   string
	PdftoppmPath string
}

func FromImage(r io.Reader) ([]byte, error) {
	src, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	thumb := imaging.Fit(src, MaxDimension, MaxDimension, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, thumb, imaging.JPEG, imaging.JPEGQuality(jpegQuality)); err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}
	return buf.Bytes(), nil
}

func FromVideo(srcPath, ffmpegPath string) ([]byte, error) {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	if _, err := exec.LookPath(ffmpegPath); err != nil {
		return nil, ErrToolUnavailable
	}

	ctx, cancel := context.WithTimeout(context.Background(), externalTimeout)
	defer cancel()

	scale := fmt.Sprintf("scale=w=%d:h=%d:force_original_aspect_ratio=decrease", MaxDimension, MaxDimension)
	cmd := exec.CommandContext(ctx, ffmpegPath,
		"-ss", fmt.Sprintf("%d", videoSeekSeconds),
		"-i", srcPath,
		"-frames:v", "1",
		"-vf", scale,
		"-f", "mjpeg",
		"-q:v", "3",
		"pipe:1",
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w", err)
	}
	if out.Len() == 0 {
		return nil, errors.New("ffmpeg produced empty output")
	}
	return out.Bytes(), nil
}

func FromPDF(srcPath, pdftoppmPath string) ([]byte, error) {
	if pdftoppmPath == "" {
		pdftoppmPath = "pdftoppm"
	}
	if _, err := exec.LookPath(pdftoppmPath); err != nil {
		return nil, ErrToolUnavailable
	}

	ctx, cancel := context.WithTimeout(context.Background(), externalTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, pdftoppmPath,
		"-jpeg",
		"-f", "1",
		"-l", "1",
		"-scale-to", fmt.Sprintf("%d", MaxDimension),
		"-singlefile",
		srcPath,
		"-",
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %w", err)
	}
	if out.Len() == 0 {
		return nil, errors.New("pdftoppm produced empty output")
	}
	return out.Bytes(), nil
}
