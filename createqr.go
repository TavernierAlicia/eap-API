package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/spf13/viper"
)

func createQR(token string, context bool) (err error) {

	var link string
	if context {
		link = "http://dashboard.easy-as-pie.fr/bartender/" + token
	} else {
		link = "http://app.easy-as-pie.fr/qr/" + token
	}
	// Create the QRcode
	qrCode, _ := qr.Encode(link, qr.M, qr.Auto)

	// Scale image
	qrCode, _ = barcode.Scale(qrCode, 200, 200)

	// create the output file
	var file *os.File
	if context {

		file, _ = os.Create(viper.GetString("links.cdn_qr_dest") + "bartender/" + fmt.Sprintf("%v", token) + ".png")
	} else {
		file, _ = os.Create(viper.GetString("links.cdn_qr_dest") + "menu_qr/" + fmt.Sprintf("%v", token) + ".png")
	}
	defer file.Close()

	// encode the barcode as png
	png.Encode(file, qrCode)

	return err
}
