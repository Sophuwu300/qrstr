/*
- Example using qrstr to write a QR code to the terminal using xterm control sequences.
- When run in an xterm-compatible terminal, it will display the QR code directly in the terminal.
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
	qr, err := qrstr.NewEncoder(qrstr.TerminalMode, qrstr.ErrorCorrection7Percent)
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
