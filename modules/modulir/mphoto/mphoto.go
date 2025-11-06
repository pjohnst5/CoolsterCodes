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

// OptimizationStats tracks space savings from image optimization
type OptimizationStats struct {
	TotalOriginalSize  int64
	TotalOptimizedSize int64
	FilesProcessed     int
	FilesOptimized     int
	FilesSkipped       int
}

// GetSpaceSaved returns the space saved in bytes
func (s *OptimizationStats) GetSpaceSaved() int64 {
	return s.TotalOriginalSize - s.TotalOptimizedSize
}

// GetCompressionRatio returns the compression ratio as a percentage
func (s *OptimizationStats) GetCompressionRatio() float64 {
	if s.TotalOriginalSize == 0 {
		return 0
	}
	return (1.0 - float64(s.TotalOptimizedSize)/float64(s.TotalOriginalSize)) * 100.0
}

// OptimizationOptions contains settings for image optimization
type OptimizationOptions struct {
	MaxWidth    int // Maximum width for images (px)
	MaxHeight   int // Maximum height for images (px)
	JpegQuality int // JPEG quality (0-100)
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
	c.Log.Infof("=== Phase 1: Optimizing original images in-place ===")
	stats, err := optimizeImagesInPlace(c, source, opts)
	if err != nil {
		return err
	}

	// Report optimization statistics
	reportOptimizationStats(c, stats, source)

	// Second pass: Copy the now-optimized images normally
	c.Log.Infof("=== Phase 2: Copying optimized images ===")
	err = copyOptimizedImages(c, source, target)
	if err != nil {
		return err
	}

	c.Log.Infof("=== Image processing complete ===")
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
func optimizeImagesInPlace(c *modulir.Context, source string, opts *OptimizationOptions) (*OptimizationStats, error) {
	stats := &OptimizationStats{}

	dirs, err := mfile.ReadDirWithOptions(c, source, &mfile.ReadDirOptions{ShowDirs: true})
	if err != nil {
		return stats, err
	}

	for _, dir := range dirs {
		// Read the files from that dir ignoring *.md
		files, err := mfile.ReadDirWithOptions(c, dir, &mfile.ReadDirOptions{IgnoreMDs: true})
		if err != nil {
			return stats, err
		}

		// Optimize all image files in-place
		for _, file := range files {
			if err = optimizeImageInPlace(c, file, stats, opts); err != nil {
				c.Log.Errorf("Error optimizing image %s: %v", file, err)
				// Continue with other files even if one fails
				continue
			}
		}
	}

	return stats, nil
}

// copyOptimizedImages recursively copies all files from source to target normally
// (assumes images have already been optimized in-place)
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
// and tracks statistics about the optimization
func optimizeImageInPlace(c *modulir.Context, imagePath string, stats *OptimizationStats, opts *OptimizationOptions) error {
	fileName := filepath.Base(imagePath)

	// Get original file size
	originalInfo, err := os.Stat(imagePath)
	if err != nil {
		return xerrors.Errorf("error getting file info for %s: %w", imagePath, err)
	}
	originalSize := originalInfo.Size()
	stats.TotalOriginalSize += originalSize
	stats.FilesProcessed++

	// Check if it's an image file
	if !isImageFile(imagePath) {
		// Not an image, skip optimization
		stats.TotalOptimizedSize += originalSize
		stats.FilesSkipped++
		return nil
	}

	// Open and decode the image
	src, err := imaging.Open(imagePath)
	if err != nil {
		// If we can't open it as an image, skip optimization
		c.Log.Debugf("Could not decode %s as image, skipping: %v", imagePath, err)
		stats.TotalOptimizedSize += originalSize
		stats.FilesSkipped++
		return nil
	}

	originalBounds := src.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	c.Log.Debugf("Processing image %s (%dx%d, %s)", fileName, originalWidth, originalHeight, formatBytes(originalSize))

	// Check if resizing is needed
	if originalWidth <= opts.MaxWidth && originalHeight <= opts.MaxHeight {
		// Image is already small enough, no optimization needed
		stats.TotalOptimizedSize += originalSize
		stats.FilesSkipped++
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
		return saveErr
	}

	// Get final file size and update stats
	if finalInfo, err := os.Stat(imagePath); err == nil {
		finalSize := finalInfo.Size()
		stats.TotalOptimizedSize += finalSize

		savings := originalSize - finalSize
		compressionPercent := float64(savings) / float64(originalSize) * 100.0

		c.Log.Infof("Optimized %s in-place: %s -> %s (saved %s, %.1f%% compression)",
			fileName, formatBytes(originalSize), formatBytes(finalSize),
			formatBytes(savings), compressionPercent)
	}

	stats.FilesOptimized++
	return nil
}

// isImageFile checks if a file is likely an image based on its extension
// Only handles JPG/JPEG and PNG files
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
// while maintaining the original aspect ratio
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

// formatBytes formats bytes into a human-readable string
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

// reportOptimizationStats logs a summary of the image optimization results
func reportOptimizationStats(c *modulir.Context, stats *OptimizationStats, sourceDir string) {
	if stats.FilesProcessed == 0 {
		c.Log.Infof("No files processed in %s", sourceDir)
		return
	}

	spaceSaved := stats.GetSpaceSaved()
	compressionRatio := stats.GetCompressionRatio()

	c.Log.Infof("=== Image Optimization Summary for %s ===", sourceDir)
	c.Log.Infof("Files processed: %d", stats.FilesProcessed)
	c.Log.Infof("Files optimized: %d", stats.FilesOptimized)
	c.Log.Infof("Files skipped: %d", stats.FilesSkipped)
	c.Log.Infof("Original total size: %s", formatBytes(stats.TotalOriginalSize))
	c.Log.Infof("Optimized total size: %s", formatBytes(stats.TotalOptimizedSize))
	c.Log.Infof("Space saved: %s (%.1f%% compression)", formatBytes(spaceSaved), compressionRatio)

	if stats.FilesOptimized > 0 {
		avgSavingsPerFile := spaceSaved / int64(stats.FilesOptimized)
		c.Log.Infof("Average savings per optimized file: %s", formatBytes(avgSavingsPerFile))
	}
}
