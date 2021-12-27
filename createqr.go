package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

func CreateQR(link string, etabId int64) (err error) {

	// Create the QRcode
	qrCode, _ := qr.Encode(link, qr.M, qr.Auto)

	// Scale image
	qrCode, _ = barcode.Scale(qrCode, 200, 200)

	// create the output file
	file, _ := os.Create("./qr/" + fmt.Sprintf("%v", etabId) + ".png")
	defer file.Close()

	// encode the barcode as png
	png.Encode(file, qrCode)

	return err
}
