/*
- Example: using qrstr to generate a QR code as plain text

Plain text QR codes have different encoders for light and dark backgrounds,
So you need to choose the appropriate encoder based on your background color.

You will need use a monospaced font to display the QR code correctly.
Plain text requires fixed-width characters to align the QR code properly.

This example makes a QR code for a dark background, with white text.
*/
package main

import (
	"fmt"

	"git.sophuwu.com/qrstr"
)

var (

	// data is the string to be encoded in the QR code
	data = "https://git.sophuwu.com/qrstr"

	// If provided, headers will be displayed as text above the QR code.
	// headers intended as human-readable labels, titles or descriptions.
	// The headers will be displayed in the order they are provided.
	headers = []string{"qrstr: QR text encoder", "scan for git repo"}
)

func main() {
	qr, err := qrstr.NewEncoder(qrstr.TextDarkMode, qrstr.ErrorCorrection7Percent)
	if err != nil {
		panic(err)
	}
	var out string
	out, err = qr.Encode(data, headers...)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
