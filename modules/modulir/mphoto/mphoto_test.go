package mphoto

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/disintegration/imaging"

	"coolstercodes/modules/modulir"
)

// createTestContext creates a test context for modulir
func createTestContext(_ *testing.T) *modulir.Context {
	return modulir.NewContext(&modulir.Args{
		Log: &modulir.Logger{Level: modulir.LevelInfo},
	})
}

// createTestImage creates a test image with specified dimensions and saves it to the given path
func createTestImage(t *testing.T, path string, width, height int, format string) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a gradient pattern for testing
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	// Create the file
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test image file %s: %v", path, err)
	}
	defer file.Close()

	// Save in the specified format
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(file, img)
	default:
		t.Fatalf("Unsupported test image format: %s", format)
	}

	if err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}
}

// getImageDimensions returns the dimensions of an image file
func getImageDimensions(t *testing.T, path string) (int, int) {
	img, err := imaging.Open(path)
	if err != nil {
		t.Fatalf("Failed to open image %s: %v", path, err)
	}
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy()
}

func TestIsImageFile(t *testing.T) {
	testCases := []struct {
		filename string
		expected bool
	}{
		{"test.jpg", true},
		{"test.jpeg", true},
		{"test.png", true},
		{"test.gif", false},  // No longer supported
		{"test.bmp", false},  // No longer supported
		{"test.webp", false}, // No longer supported
		{"test.heic", false}, // No longer supported
		{"test.JPG", true},   // Case insensitive
		{"test.PNG", true},
		{"test.txt", false},
		{"test.md", false},
		{"test", false},
		{"test.pdf", false},
	}

	for _, tc := range testCases {
		result := isImageFile(tc.filename)
		if result != tc.expected {
			t.Errorf("isImageFile(%s) = %v, want %v", tc.filename, result, tc.expected)
		}
	}
}

func TestCalculateNewDimensions(t *testing.T) {
	testCases := []struct {
		name                          string
		originalWidth, originalHeight int
		maxWidth, maxHeight           int
		expectedWidth, expectedHeight int
	}{
		{
			name:           "No resize needed",
			originalWidth:  800,
			originalHeight: 600,
			maxWidth:       1200,
			maxHeight:      1200,
			expectedWidth:  800,
			expectedHeight: 600,
		},
		{
			name:           "Width exceeds limit",
			originalWidth:  2000,
			originalHeight: 1000,
			maxWidth:       1200,
			maxHeight:      1200,
			expectedWidth:  1200,
			expectedHeight: 600,
		},
		{
			name:           "Height exceeds limit",
			originalWidth:  800,
			originalHeight: 1500,
			maxWidth:       1200,
			maxHeight:      1200,
			expectedWidth:  640,
			expectedHeight: 1200,
		},
		{
			name:           "Both dimensions exceed limit",
			originalWidth:  2400,
			originalHeight: 1800,
			maxWidth:       1200,
			maxHeight:      1200,
			expectedWidth:  1200,
			expectedHeight: 900,
		},
		{
			name:           "Portrait image",
			originalWidth:  1000,
			originalHeight: 2000,
			maxWidth:       1200,
			maxHeight:      1200,
			expectedWidth:  600,
			expectedHeight: 1200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			width, height := calculateNewDimensions(
				tc.originalWidth, tc.originalHeight,
				tc.maxWidth, tc.maxHeight,
			)
			if width != tc.expectedWidth || height != tc.expectedHeight {
				t.Errorf("calculateNewDimensions(%d, %d, %d, %d) = (%d, %d), want (%d, %d)",
					tc.originalWidth, tc.originalHeight, tc.maxWidth, tc.maxHeight,
					width, height, tc.expectedWidth, tc.expectedHeight)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1536 * 1024, "1.5 MB"},
		{2147483648, "2.0 GB"},
	}

	for _, tc := range testCases {
		result := formatBytes(tc.bytes)
		if result != tc.expected {
			t.Errorf("formatBytes(%d) = %s, want %s", tc.bytes, result, tc.expected)
		}
	}
}

func TestOptimizeImageInPlace(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	testCases := []struct {
		name           string
		filename       string
		width          int
		height         int
		format         string
		maxWidth       int
		maxHeight      int
		shouldOptimize bool
	}{
		{
			name:           "Large JPEG should be optimized",
			filename:       "large.jpg",
			width:          2000,
			height:         1500,
			format:         "jpeg",
			maxWidth:       1200,
			maxHeight:      1200,
			shouldOptimize: true,
		},
		{
			name:           "Small image should not be optimized",
			filename:       "small.jpg",
			width:          800,
			height:         600,
			format:         "jpeg",
			maxWidth:       1200,
			maxHeight:      1200,
			shouldOptimize: false,
		},
		{
			name:           "Large PNG should be optimized",
			filename:       "large.png",
			width:          1600,
			height:         1200,
			format:         "png",
			maxWidth:       1200,
			maxHeight:      1200,
			shouldOptimize: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test image
			imagePath := filepath.Join(tempDir, tc.filename)
			createTestImage(t, imagePath, tc.width, tc.height, tc.format)

			// Get original dimensions
			originalWidth, originalHeight := getImageDimensions(t, imagePath)

			// Create test context and options
			ctx := createTestContext(t)
			opts := &OptimizationOptions{
				MaxWidth:    tc.maxWidth,
				MaxHeight:   tc.maxHeight,
				JpegQuality: 85,
			}

			// Run optimization
			err := optimizeImageInPlace(ctx, imagePath, opts)
			if err != nil {
				t.Fatalf("optimizeImageInPlace failed: %v", err)
			}

			// Verify dimensions after optimization
			newWidth, newHeight := getImageDimensions(t, imagePath)

			if tc.shouldOptimize {
				// Should be resized
				if newWidth > tc.maxWidth || newHeight > tc.maxHeight {
					t.Errorf("Image not properly resized: got %dx%d, max allowed %dx%d",
						newWidth, newHeight, tc.maxWidth, tc.maxHeight)
				}

				// Original aspect ratio should be maintained (within 1% tolerance)
				originalAspect := float64(originalWidth) / float64(originalHeight)
				newAspect := float64(newWidth) / float64(newHeight)
				aspectDiff := (originalAspect - newAspect) / originalAspect
				if aspectDiff > 0.01 || aspectDiff < -0.01 {
					t.Errorf("Aspect ratio not maintained: original %.3f, new %.3f",
						originalAspect, newAspect)
				}
			} else {
				// Should not be resized
				if newWidth != originalWidth || newHeight != originalHeight {
					t.Errorf("Small image was unexpectedly resized: %dx%d -> %dx%d",
						originalWidth, originalHeight, newWidth, newHeight)
				}
			}
		})
	}
}

func TestOptimizeImageInPlace_NonImageFile(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create a text file
	textPath := filepath.Join(tempDir, "test.txt")
	content := []byte("This is a test file")
	if err := os.WriteFile(textPath, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create test context and options
	ctx := createTestContext(t)
	opts := &OptimizationOptions{
		MaxWidth:    1200,
		MaxHeight:   1200,
		JpegQuality: 85,
	}

	// Run optimization on non-image file (should be skipped silently)
	err := optimizeImageInPlace(ctx, textPath, opts)
	if err != nil {
		t.Fatalf("optimizeImageInPlace failed on non-image file: %v", err)
	}

	// Verify file content unchanged
	newContent, err := os.ReadFile(textPath)
	if err != nil {
		t.Fatalf("Failed to read file after optimization: %v", err)
	}
	if string(newContent) != string(content) {
		t.Errorf("Non-image file content was modified")
	}
}

func TestOptimizeImageInPlace_CorruptImage(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create a corrupt "image" file
	corruptPath := filepath.Join(tempDir, "corrupt.jpg")
	corruptContent := []byte("This is not a valid JPEG file but has .jpg extension")
	if err := os.WriteFile(corruptPath, corruptContent, 0644); err != nil {
		t.Fatalf("Failed to create corrupt test file: %v", err)
	}

	// Create test context and options
	ctx := createTestContext(t)
	opts := &OptimizationOptions{
		MaxWidth:    1200,
		MaxHeight:   1200,
		JpegQuality: 85,
	}

	// Run optimization on corrupt image file (should be skipped silently)
	err := optimizeImageInPlace(ctx, corruptPath, opts)
	if err != nil {
		t.Fatalf("optimizeImageInPlace failed on corrupt image: %v", err)
	}

	// Verify file content unchanged (since it couldn't be processed)
	newContent, err := os.ReadFile(corruptPath)
	if err != nil {
		t.Fatalf("Failed to read file after optimization: %v", err)
	}
	if string(newContent) != string(corruptContent) {
		t.Errorf("Corrupt image file content was modified")
	}
}

// Benchmark tests for performance evaluation
func BenchmarkCalculateNewDimensions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calculateNewDimensions(2000, 1500, 1200, 1200)
	}
}

func BenchmarkIsImageFile(b *testing.B) {
	filenames := []string{
		"test.jpg", "test.png", "test.gif", "test.txt", "test.md",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, filename := range filenames {
			isImageFile(filename)
		}
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	sizes := []int64{0, 1024, 1048576, 1073741824}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, size := range sizes {
			formatBytes(size)
		}
	}
}
