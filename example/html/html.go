/*
- Example using qrstr to generate a QR code in HTML format.
- When run, it will print the HTML representation of the QR code to the console.
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
	qr, err := qrstr.NewEncoder(qrstr.HTMLMode, qrstr.ErrorCorrection25Percent)
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
