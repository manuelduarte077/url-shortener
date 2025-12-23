package qrcode

import (
	"testing"

	"github.com/yeqown/go-qrcode/v2"
)

func TestNewQRCodeGenerator(t *testing.T) {
	gen := NewQRCodeGenerator()
	if gen == nil {
		t.Fatal("NewQRCodeGenerator returned nil")
	}
	if gen.Size != 512 {
		t.Errorf("expected default size 512, got %d", gen.Size)
	}
	if gen.ErrorCorrectionLevel != qrcode.ErrorCorrectionHighest {
		t.Errorf("expected default error correction level Highest, got %v", gen.ErrorCorrectionLevel)
	}
}

func TestGeneratePNG(t *testing.T) {
	gen := NewQRCodeGenerator()
	testURL := "https://example.com/test"

	pngData, err := gen.GeneratePNG(testURL)
	if err != nil {
		t.Fatalf("GeneratePNG failed: %v", err)
	}

	if len(pngData) == 0 {
		t.Fatal("GeneratePNG returned empty data")
	}

	if len(pngData) < 8 {
		t.Fatal("PNG data too short")
	}
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i := 0; i < 8; i++ {
		if pngData[i] != pngSignature[i] {
			t.Fatalf("invalid PNG signature at byte %d", i)
		}
	}
}

func TestGeneratePNGWithOptions(t *testing.T) {
	gen := NewQRCodeGenerator()
	testURL := "https://example.com/test"

	pngData, err := gen.GeneratePNGWithOptions(testURL)
	if err != nil {
		t.Fatalf("GeneratePNGWithOptions failed: %v", err)
	}

	if len(pngData) == 0 {
		t.Fatal("GeneratePNGWithOptions returned empty data")
	}
}

func TestGeneratePNG_EmptyURL(t *testing.T) {
	gen := NewQRCodeGenerator()

	_, err := gen.GeneratePNG("")
	if err == nil {
		t.Error("expected error for empty URL, got nil")
	}
}

func TestGeneratePNG_DifferentSizes(t *testing.T) {
	testURL := "https://example.com/test"

	sizes := []int{128, 256, 512}
	for _, size := range sizes {
		gen := NewQRCodeGenerator()
		gen.Size = size

		pngData, err := gen.GeneratePNG(testURL)
		if err != nil {
			t.Fatalf("GeneratePNG failed for size %d: %v", size, err)
		}

		if len(pngData) == 0 {
			t.Fatalf("GeneratePNG returned empty data for size %d", size)
		}
	}
}
