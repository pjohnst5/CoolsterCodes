package main

import (
	"fmt"
	"testing"
)

func TestImageDimensionCalculation(t *testing.T) {
	testCases := []struct {
		originalW, originalH int
		expectedW, expectedH int
		description         string
	}{
		{1200, 800, 800, 533, "Wide image (1200x800)"},
		{600, 900, 533, 800, "Tall image (600x900)"},
		{400, 300, 400, 300, "Small image (400x300)"},
		{1600, 1200, 800, 600, "Large square-ish image (1600x1200)"},
		{2000, 500, 800, 200, "Very wide banner (2000x500)"},
	}
	
	for _, tc := range testCases {
		newW, newH := calculateNewDimensions(tc.originalW, tc.originalH, MaxImageWidth, MaxImageHeight)
		fmt.Printf("%s: %dx%d -> %dx%d (expected %dx%d)\n", 
			tc.description, tc.originalW, tc.originalH, newW, newH, tc.expectedW, tc.expectedH)
		
		if newW != tc.expectedW || newH != tc.expectedH {
			t.Errorf("Dimension calculation failed for %s: got %dx%d, expected %dx%d", 
				tc.description, newW, newH, tc.expectedW, tc.expectedH)
		}
	}
}

func TestIsImageFile(t *testing.T) {
	testFiles := []struct {
		filename string
		expected bool
	}{
		{"image.jpg", true},
		{"photo.jpeg", true},
		{"graphic.png", true},
		{"animation.gif", true},
		{"icon.webp", true},
		{"picture.heic", true},
		{"document.pdf", false},
		{"text.txt", false},
		{"data.json", false},
		{"style.css", false},
	}
	
	for _, tc := range testFiles {
		result := isImageFile(tc.filename)
		fmt.Printf("File %s: %t (expected %t)\n", tc.filename, result, tc.expected)
		
		if result != tc.expected {
			t.Errorf("File type detection failed for %s: got %t, expected %t", 
				tc.filename, result, tc.expected)
		}
	}
}