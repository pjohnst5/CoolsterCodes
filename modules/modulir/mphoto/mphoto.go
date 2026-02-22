package mphoto

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/xerrors"

	"coolstercodes/modules/modulir"
	"coolstercodes/modules/modulir/mfile"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Types and Constants
//
//
//
//////////////////////////////////////////////////////////////////////////////

// OptimizationOptions contains settings for image optimization.
type OptimizationOptions struct {
	MaxWidth        int   // Maximum width for images (px)
	MaxHeight       int   // Maximum height for images (px)
	JpegQuality     int   // JPEG quality (0-100)
	MaxFileSizeBytes int64 // Maximum file size in bytes (0 means no limit)
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Public Functions
//
//
//
//////////////////////////////////////////////////////////////////////////////

// CopyDirectoryImagesOptimized first optimizes all original images in-place,
// then copies the optimized images from source to target.
func CopyDirectoryImagesOptimized(c *modulir.Context, source, target string, opts *OptimizationOptions) error {
	// First pass: Optimize all original images in-place
	err := optimizeImagesInPlace(c, source, opts)
	if err != nil {
		return err
	}

	// Second pass: Copy the now-optimized images
	err = copyOptimizedImages(c, source, target)
	if err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private Functions
//
//
//
//////////////////////////////////////////////////////////////////////////////

// optimizeImagesInPlace recursively optimizes all images in the source directory in-place,
// resizing them if they exceed the maximum dimensions while maintaining aspect ratio.
func optimizeImagesInPlace(c *modulir.Context, source string, opts *OptimizationOptions) error {
	dirs, err := mfile.ReadDirWithOptions(c, source, &mfile.ReadDirOptions{ShowDirs: true})
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		// Read the files from that dir ignoring *.md
		files, err := mfile.ReadDirWithOptions(c, dir, &mfile.ReadDirOptions{IgnoreMDs: true})
		if err != nil {
			return err
		}

		// Optimize all image files in-place
		for _, file := range files {
			if err = optimizeImageInPlace(c, file, opts); err != nil {
				c.Log.Errorf("Error optimizing image %s: %v", file, err)
				// Continue with other files even if one fails
				continue
			}
		}
	}

	return nil
}

// copyOptimizedImages recursively copies all files from source to target normally
// (assumes images have already been optimized in-place).
func copyOptimizedImages(c *modulir.Context, source, target string) error {
	dirs, err := mfile.ReadDirWithOptions(c, source, &mfile.ReadDirOptions{ShowDirs: true})
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		// Read the files from that dir ignoring *.md
		files, err := mfile.ReadDirWithOptions(c, dir, &mfile.ReadDirOptions{IgnoreMDs: true})
		if err != nil {
			return err
		}

		// Make new target directory path
		justNameOfDir := filepath.Base(dir)
		targetDir := filepath.Join(target, justNameOfDir)

		// Ensure target directory exists
		if err = mfile.EnsureDir(c, targetDir); err != nil {
			return err
		}

		// Copy all files normally (no optimization needed since already done)
		for _, file := range files {
			fileName := filepath.Base(file)
			targetPath := filepath.Join(targetDir, fileName)

			if err = mfile.CopyFile(c, file, targetPath); err != nil {
				c.Log.Errorf("Error copying file %s: %v", file, err)
				// Continue with other files even if one fails
				continue
			}
		}
	}

	return nil
}

// optimizeImageInPlace optimizes a single image file in-place, resizing it if necessary,
// and tracks statistics about the optimization.
func optimizeImageInPlace(c *modulir.Context, imagePath string, opts *OptimizationOptions) error {
	fileName := filepath.Base(imagePath)

	// Get original file size
	originalInfo, err := os.Stat(imagePath)
	if err != nil {
		return xerrors.Errorf("error getting file info for %s: %w", imagePath, err)
	}
	originalSize := originalInfo.Size()

	// Check if it's an image file
	if !isImageFile(imagePath) {
		return nil
	}

	// Open and decode the image
	src, err := imaging.Open(imagePath)
	if err != nil {
		// Skip files that can't be decoded as images
		c.Log.Debugf("Could not decode %s as image, skipping: %v", imagePath, err)
		return nil
	}

	originalBounds := src.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	c.Log.Debugf("Processing image %s (%dx%d, %s)", fileName, originalWidth, originalHeight, formatBytes(originalSize))

	// Check if file size exceeds the limit
	if opts.MaxFileSizeBytes > 0 && originalSize > opts.MaxFileSizeBytes {
		return xerrors.Errorf("image %s exceeds maximum file size: %s > %s",
			fileName, formatBytes(originalSize), formatBytes(opts.MaxFileSizeBytes))
	}

	// Check if resizing is needed
	if originalWidth <= opts.MaxWidth && originalHeight <= opts.MaxHeight {
		return nil
	}

	// Calculate new dimensions while maintaining aspect ratio
	newWidth, newHeight := calculateNewDimensions(originalWidth, originalHeight, opts.MaxWidth, opts.MaxHeight)

	c.Log.Infof("Resizing image %s from %dx%d to %dx%d (in-place)", fileName, originalWidth, originalHeight, newWidth, newHeight)

	// Resize the image
	resized := imaging.Resize(src, newWidth, newHeight, imaging.Lanczos)

	// Save the resized image back to the original path (in-place optimization)
	ext := strings.ToLower(filepath.Ext(imagePath))
	var saveErr error
	switch ext {
	case ".jpg", ".jpeg":
		saveErr = imaging.Save(resized, imagePath, imaging.JPEGQuality(opts.JpegQuality))
	case ".png":
		saveErr = imaging.Save(resized, imagePath)
	default:
		// This should not happen since isImageFile() only allows JPG/JPEG and PNG
		return xerrors.Errorf("unsupported image format: %s", ext)
	}

	if saveErr != nil {
		return xerrors.Errorf("failed to save optimized image %s: %w", imagePath, saveErr)
	}

	// Get final file size and update stats
	if finalInfo, err := os.Stat(imagePath); err == nil {
		finalSize := finalInfo.Size()

		savings := originalSize - finalSize
		compressionPercent := float64(savings) / float64(originalSize) * 100.0

		c.Log.Infof("Optimized %s in-place: %s -> %s (saved %s, %.1f%% compression)",
			fileName, formatBytes(originalSize), formatBytes(finalSize),
			formatBytes(savings), compressionPercent)
	}
	return nil
}

// isImageFile checks if a file is likely an image based on its extension.
// Only handles JPG/JPEG and PNG files.
func isImageFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	imageExts := []string{".jpg", ".jpeg", ".png"}

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// calculateNewDimensions calculates new width and height that fit within maxWidth and maxHeight
// while maintaining the original aspect ratio.
func calculateNewDimensions(originalWidth, originalHeight, maxWidth, maxHeight int) (int, int) {
	if originalWidth <= maxWidth && originalHeight <= maxHeight {
		return originalWidth, originalHeight
	}

	// Calculate scaling factors
	widthScale := float64(maxWidth) / float64(originalWidth)
	heightScale := float64(maxHeight) / float64(originalHeight)

	// Use the smaller scaling factor to ensure both dimensions fit
	scale := widthScale
	if heightScale < widthScale {
		scale = heightScale
	}

	newWidth := int(float64(originalWidth) * scale)
	newHeight := int(float64(originalHeight) * scale)

	return newWidth, newHeight
}

// formatBytes formats bytes into a human-readable string.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
