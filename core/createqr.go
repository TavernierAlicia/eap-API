package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

func createQR(token string, context bool) (err error) {

	var link string
	if context {
		link = "http://localhost:9999/bartender/" + token
	} else {
		link = "http://localhost:9999/menu/" + token
	}
	// Create the QRcode
	qrCode, _ := qr.Encode(link, qr.M, qr.Auto)

	// Scale image
	qrCode, _ = barcode.Scale(qrCode, 200, 200)

	// create the output file
	if context {
		file, _ := os.Create("../media/qrs/bartender/" + fmt.Sprintf("%v", token) + ".png")
	} else {
		file, _ := os.Create("../media/qrs/menu_qr/" + fmt.Sprintf("%v", token) + ".png")
	}
	defer file.Close()

	// encode the barcode as png
	png.Encode(file, qrCode)

	return err
}
