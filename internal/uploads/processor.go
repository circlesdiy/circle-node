package uploads

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"

	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

const (
	// Avatar dimensions
	AvatarWidth  = 512
	AvatarHeight = 512

	// Banner dimensions
	BannerWidth  = 1500
	BannerHeight = 500

	// Max dimensions to prevent decompression bombs
	MaxImageWidth  = 10000
	MaxImageHeight = 10000

	// JPEG quality for encoding
	JPEGQuality = 85
)

// ImageProcessor handles image processing operations
type ImageProcessor struct{}

// NewImageProcessor creates a new image processor
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

// ProcessAvatar decodes, validates, resizes and re-encodes an avatar image
// Returns a buffer containing the processed image as JPEG
func (p *ImageProcessor) ProcessAvatar(file multipart.File) (*bytes.Buffer, error) {
	return p.processImage(file, AvatarWidth, AvatarHeight)
}

// ProcessBanner decodes, validates, resizes and re-encodes a banner image
// Returns a buffer containing the processed image as JPEG
func (p *ImageProcessor) ProcessBanner(file multipart.File) (*bytes.Buffer, error) {
	return p.processImage(file, BannerWidth, BannerHeight)
}

// processImage is the core image processing function
func (p *ImageProcessor) processImage(file multipart.File, targetWidth, targetHeight int) (*bytes.Buffer, error) {
	// Decode image (supports JPEG, PNG, WebP)
	img, format, err := image.Decode(file)
	if err != nil {
		// Try WebP if standard decode fails
		if _, seekErr := file.Seek(0, 0); seekErr != nil {
			return nil, fmt.Errorf("failed to reset file pointer: %w", seekErr)
		}
		img, err = webp.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image: %w", err)
		}
		format = "webp"
	}

	// Validate dimensions to prevent decompression bombs
	bounds := img.Bounds()
	if bounds.Dx() > MaxImageWidth || bounds.Dy() > MaxImageHeight {
		return nil, fmt.Errorf("image dimensions %dx%d exceed maximum %dx%d",
			bounds.Dx(), bounds.Dy(), MaxImageWidth, MaxImageHeight)
	}

	// Resize image to target dimensions
	// Using Lanczos resampling for high quality
	resized := imaging.Fill(img, targetWidth, targetHeight, imaging.Center, imaging.Lanczos)

	// Encode as JPEG (automatically strips EXIF and other metadata)
	var buf bytes.Buffer
	if format == "png" {
		// For PNG, maintain transparency if present, otherwise use JPEG
		if hasTransparency(img) {
			err = png.Encode(&buf, resized)
		} else {
			err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: JPEGQuality})
		}
	} else {
		// For JPEG and WebP, encode as JPEG
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: JPEGQuality})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return &buf, nil
}

// hasTransparency checks if an image has transparency
func hasTransparency(img image.Image) bool {
	switch i := img.(type) {
	case *image.NRGBA:
		// Check if any pixel has alpha < 255
		bounds := i.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, a := i.At(x, y).RGBA()
				if a < 65535 { // RGBA returns values in range [0, 65535]
					return true
				}
			}
		}
	case *image.RGBA:
		bounds := i.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, a := i.At(x, y).RGBA()
				if a < 65535 {
					return true
				}
			}
		}
	}
	return false
}
