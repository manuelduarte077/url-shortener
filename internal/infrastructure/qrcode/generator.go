package qrcode

import (
	"bytes"
	"fmt"
	"image/png"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

type QRCodeGenerator struct {
	Size                 int
	ErrorCorrectionLevel interface{}
}

func NewQRCodeGenerator() *QRCodeGenerator {
	return &QRCodeGenerator{
		Size:                 512,                           // Increased for better quality
		ErrorCorrectionLevel: qrcode.ErrorCorrectionHighest, // Highest error correction for better quality
	}
}

type writeCloser struct {
	*bytes.Buffer
}

func (wc *writeCloser) Close() error {
	return nil
}

func (g *QRCodeGenerator) GeneratePNG(url string) ([]byte, error) {
	if url == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}

	var encodeOpt qrcode.EncodeOption
	switch g.ErrorCorrectionLevel {
	case qrcode.ErrorCorrectionLow:
		encodeOpt = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionLow)
	case qrcode.ErrorCorrectionQuart:
		encodeOpt = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionQuart)
	case qrcode.ErrorCorrectionHighest:
		encodeOpt = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionHighest)
	default:
		encodeOpt = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium)
	}

	qrc, err := qrcode.NewWith(url, encodeOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	var buf bytes.Buffer
	wc := &writeCloser{Buffer: &buf}

	width := uint8(25)
	if g.Size > 0 {
		calculatedWidth := g.Size / 20
		if calculatedWidth < 5 {
			width = 5
		} else if calculatedWidth > 50 {
			width = 50
		} else {
			width = uint8(calculatedWidth)
		}
	}

	w := standard.NewWithWriter(wc,
		standard.WithQRWidth(width),
		standard.WithBorderWidth(2),
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
	)

	if err := qrc.Save(w); err != nil {
		return nil, fmt.Errorf("failed to save QR code: %w", err)
	}

	if err := wc.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (g *QRCodeGenerator) GeneratePNGWithOptions(
	url string,
	opts ...qrcode.EncodeOption,
) ([]byte, error) {
	var ecOpt qrcode.EncodeOption
	switch g.ErrorCorrectionLevel {
	case qrcode.ErrorCorrectionLow:
		ecOpt = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionLow)
	case qrcode.ErrorCorrectionQuart:
		ecOpt = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionQuart)
	case qrcode.ErrorCorrectionHighest:
		ecOpt = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionHighest)
	default:
		ecOpt = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium)
	}

	baseOpts := []qrcode.EncodeOption{ecOpt}
	allOpts := append(baseOpts, opts...)

	qrc, err := qrcode.NewWith(url, allOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	var buf bytes.Buffer
	wc := &writeCloser{Buffer: &buf}

	width := uint8(25)
	if g.Size > 0 {
		calculatedWidth := g.Size / 20
		if calculatedWidth < 5 {
			width = 5
		} else if calculatedWidth > 50 {
			width = 50
		} else {
			width = uint8(calculatedWidth)
		}
	}

	w := standard.NewWithWriter(wc,
		standard.WithQRWidth(width),
		standard.WithBorderWidth(2),
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
	)

	if err := qrc.Save(w); err != nil {
		return nil, fmt.Errorf("failed to save QR code: %w", err)
	}

	if err := wc.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	if _, err := png.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		return nil, fmt.Errorf("invalid PNG generated: %w", err)
	}

	return buf.Bytes(), nil
}
