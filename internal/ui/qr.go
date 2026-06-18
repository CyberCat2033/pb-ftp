package ui

import (
	"fmt"
	"image"
	"image/color"

	ink "github.com/dennwc/inkview"
	"github.com/skip2/go-qrcode"
)

type QrCodePref struct {
	scale   int
	offsetX int
	offsetY int
}

func DrawQRCenteredIn(qr *qrcode.QRCode, rect image.Rectangle) error {
	if qr == nil {
		return fmt.Errorf("QrCode is empty")
	}
	matrix := qr.Bitmap()
	matrixSize := len(matrix)
	if matrixSize == 0 {
		return fmt.Errorf("QrCode matrix is empty")
	}

	rectWidth := rect.Dx()
	rectHeight := rect.Dy()

	// Calculate maximum scale that fits in both width and height
	scaleX := rectWidth / matrixSize
	scaleY := rectHeight / matrixSize
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}
	if scale <= 0 {
		scale = 1
	}

	qrPx := matrixSize * scale
	offsetX := rect.Min.X + (rectWidth-qrPx)/2
	offsetY := rect.Min.Y + (rectHeight-qrPx)/2

	pref := QrCodePref{
		scale:   scale,
		offsetX: offsetX,
		offsetY: offsetY,
	}
	return DrawQR(qr, &pref)
}

func DrawQRCentered(qr *qrcode.QRCode, screenSize image.Point, scale int) error {
	if qr == nil {
		return fmt.Errorf("QrCode is empty")
	}
	if scale <= 0 {
		return fmt.Errorf("scale must be a positive number")
	}

	size := len(qr.Bitmap())
	qrPx := size * scale
	if screenSize.X < qrPx || screenSize.Y < qrPx {
		return fmt.Errorf("QrCode is tooooo big")
	}
	pref := QrCodePref{
		scale:   scale,
		offsetX: (screenSize.X - qrPx) / 2,
		offsetY: (screenSize.Y - qrPx) / 2,
	}
	return DrawQR(qr, &pref)
}

func DrawQR(qr *qrcode.QRCode, pref *QrCodePref) error {
	if qr == nil {
		return fmt.Errorf("QrCode is empty")
	}
	matrix := qr.Bitmap()
	size := len(matrix)

	defaultPref := QrCodePref{
		scale:   10,
		offsetX: 100,
		offsetY: 100,
	}

	if pref == nil {
		pref = &defaultPref
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if matrix[y][x] {
				x0 := pref.offsetX + x*pref.scale
				y0 := pref.offsetY + y*pref.scale

				ink.FillArea(image.Rect(
					x0,
					y0,
					x0+pref.scale,
					y0+pref.scale,
				), color.Black)
			}
		}
	}

	return nil
}
