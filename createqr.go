package main

import (
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

func CreateQR() {
	// Create the barcode
	qrCode, _ := qr.Encode("https://github.com/boombuler/barcode", qr.M, qr.Auto)

	// Scale the barcode to 200x200 pixels
	qrCode, _ = barcode.Scale(qrCode, 200, 200)

	// create the output file
	file, _ := os.Create("./qr/qrcode.png")
	defer file.Close()

	// encode the barcode as png
	png.Encode(file, qrCode)
}
