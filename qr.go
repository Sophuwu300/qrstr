package qrstr

/*
 * This file is to turn strings into unicode or html qr codes.
 * Author: sophuwu <sophie@skisiel.com>
 * Feel free to use this code in any way you want.
 * Just call QR("string", header bool, html bool) to get a qr code.
 * The header will display the string above the qr code.
 */

import (
	"fmt"
	"github.com/boombuler/barcode/qr"
	"image"
	"image/color"
	"slices"
	"strings"
)

// utf8r aliases rune
type utf8r rune

// String returns the plain text or html representation of the rune
func (u utf8r) String(html bool) string {
	if html {
		return fmt.Sprintf("<td>&#x%x;</td>", u)
	}
	return string(u)
}

const blank rune = ' '
const upper rune = '▀'
const lower rune = '▄'
const whole rune = '█'

// pad returns a string with the given length
func pad(n int, r ...rune) string {
	if n == 0 {
		return ""
	}
	if n < 0 {
		n = 1
	}
	if len(r) == 0 {
		return strings.Repeat(string(whole), n)
	}
	return strings.Repeat(string(r[0]), n)
}

type runeCol []rune

func (c *runeCol) getRune(top, bot color.Color) rune {
	return (*c)[func() int {
		i := 0
		if top == color.Black {
			i |= 1
		}
		if bot == color.Black {
			i |= 2
		}
		return i
	}()]
}

func (c *runeCol) addRune(s *string, top, bot color.Color) {
	*s += string(c.getRune(top, bot))
}

var lightMode = runeCol{blank, upper, lower, whole}
var darkMode = runeCol{whole, lower, upper, blank}

// wrap wraps text around a newline, hard wrapping at the given width
// but trying to soft wrap if possible
func wrap(w int, s ...string) []string {
	var line = ""
	var lines, b []string
	var v string
	var i, j int
	w--
	for _, l := range s {
		if len(l) < w {
			lines = append(lines, l)
			continue
		}

		b = strings.Split(l, " ")
		for i, v = range b {
			if len(v) > w {
				for j = 0; j < len(v); j += w {
					if j+w < len(v) {
						lines = append(lines, v[j:j+w]+"-")
					} else {
						line = v[j:] + " "
					}
				}
				continue
			}
			if len(line)+len(v) < w {
				line += v + " "
			} else {
				lines = append(lines, strings.TrimSuffix(line, " "))
				line = ""
				continue
			}
			if i == len(b)-1 {
				lines = append(lines, strings.TrimSuffix(line, " "))
				line = ""
				continue
			}
		}
	}

	return slices.Clip(lines)
}

type Encoder struct {
	strFunc func(rc *runeCol, code *image.Image, headers *[]string) (string, error)
	rc      *runeCol
	errCorr ErrorCorrectionLevel
}

// Encode encodes data with configuration from NewEncoder into a qr code string.
// If headers are provided, they will be displayed above the qr code in the output.
func (q *Encoder) Encode(data string, headers ...string) (string, error) {
	strFunc := q.strFunc
	if strFunc == nil {
		return "", fmt.Errorf("encoder misconfigured, use NewEncoder when creating it")
	}
	var code image.Image
	var err error
	code, err = qr.Encode(data, qr.ErrorCorrectionLevel((*q).errCorr), qr.Auto)
	if err != nil {
		return "", err
	}
	return strFunc(q.rc, &code, &headers)
}

func text(rc *runeCol, code *image.Image, headers *[]string) (string, error) {
	if rc == nil || code == nil {
		return "", fmt.Errorf("encoder misconfigured, use NewEncoder when creating it")
	}
	var output = ""
	dx := (*code).Bounds().Dx()
	dy := (*code).Bounds().Dy()
	wr := rc.getRune(color.White, color.White)
	prefix := string(wr)
	suffix := string(wr) + "\n"

	hashead := headers != nil && len(*headers) > 0

	if hashead {
		output += fmt.Sprintln(string(whole) + pad(dx+2, upper) + string(whole))
		for _, v := range wrap(dx, *headers...) {
			v = v + pad(dx-len(v)+1, blank) + string(whole)
			v = string(whole) + string(blank) + v
			output += v + "\n"
		}

		output += string(whole) + pad(dx+2, lower) + string(whole) + "\n" + string(whole) + pad(dx+2, wr) + string(whole) + "\n"
		prefix = string(whole) + string(wr)
		suffix = string(wr) + string(whole) + "\n"
	} else {
		output += pad(dx+2, wr) + "\n"
	}

	output += prefix
	prefix = suffix + prefix

	var y, x int
	for y = 0; y < dy-dy%2; y += 2 {
		for x = 0; x < dx; x++ {
			rc.addRune(&output, (*code).At(x, y), (*code).At(x, y+1))
		}
		output += prefix
	}
	if dy%2 == 1 {
		for x = 0; x < dx; x++ {
			rc.addRune(&output, (*code).At(x, y), color.White)
		}
		output += suffix
	} else {
		output = strings.TrimSuffix(output, prefix) + suffix
	}

	if hashead {
		output += pad(dx+4, wr) + "\n"
	} else {
		output += pad(dx+2, wr) + "\n"
	}

	return output, nil
}

const css = `<style>
.qrstr-white {
	background-color: white;
	color: white;
	border-color: white;
	padding: 0.5em;
}
.qrstr-black {
	background-color: black;
	color: black;
	border-color: black;
	padding: 0.5em;
}
.qrstr-code , .qrstr-code * {
	border: 0;
	font-family: monospace;
	padding: 0;
	margin: 0;
	font-size: 10%;
	letter-spacing: 0;
	border-spacing: 0;
	border-collapse: collapse;
}
</style>
`

func html(rc *runeCol, code *image.Image, headers *[]string) (string, error) {
	var output = "<div style=\"width: min-content;background: white; color: black;  padding: 1lh;\">\n" + css
	if code == nil {
		return "", fmt.Errorf("encoder misconfigured, use NewEncoder when creating it")
	}
	hashead := headers != nil && len(*headers) > 0
	if hashead {
		for _, v := range *headers {
			output += "<p>" + v + "</p>\n"
		}
		output += "<hr>\n"
	}
	dx := (*code).Bounds().Dx()
	dy := (*code).Bounds().Dy()
	output += "<table class\"qrstr-code\" style=\"border-collapse: collapse;\">\n"
	for y := 0; y < dy; y++ {
		output += "<tr>\n"
		for x := 0; x < dx; x++ {
			if (*code).At(x, y) == color.Black {
				output += "<td class=\"qrstr-black\"></td>\n"
			} else {
				output += "<td class=\"qrstr-white\"></td>\n"
			}
		}
		output += "</tr>\n"
	}
	output += "</table></div>\n"
	return output, nil
}

type EncoderType int
type ErrorCorrectionLevel qr.ErrorCorrectionLevel

const (
	// TextDarkMode makes qr codes for printing on dark backgrounds with white text,
	// like a terminal or a screen with dark/night mode enabled.
	// This mode is not recommended for terminals, use TerminalMode instead.
	// MUST BE PRINTED/DISPLAYED USING A MONOSPACE FONT.
	TextDarkMode EncoderType = 0
	// TextLightMode makes qr codes for printing on light backgrounds with black text,
	// like white paper or a screen with light mode.
	// This mode is not recommended for terminals, use TerminalMode instead.
	// MUST BE PRINTED/DISPLAYED USING A MONOSPACE FONT.
	TextLightMode EncoderType = 1
	// HTMLMode makes qr codes for embedding in HTML documents or web pages.
	// Colours are set automatically with this mode.
	// Generates a table using HTML tags, does not require a monospace font.
	HTMLMode EncoderType = 2
	// TerminalMode makes qr codes for printing on xterm terminals with auto color.
	// Colours are set automatically with this mode.
	// MUST BE PRINTED/DISPLAYED USING A MONOSPACE FONT.
	TerminalMode EncoderType = 3

	// ErrorCorrection7Percent indicates 7% of lost data can be recovered, makes the qr code smaller
	ErrorCorrection7Percent ErrorCorrectionLevel = 0
	// ErrorCorrection15Percent indicates 15% of lost data can be recovered, default
	ErrorCorrection15Percent ErrorCorrectionLevel = 1
	// ErrorCorrection25Percent indicates 25% of lost data can be recovered, makes the qr code bigger
	ErrorCorrection25Percent ErrorCorrectionLevel = 2
	// ErrorCorrection30Percent indicates 30% of lost data can be recovered, makes the qr code very big
	ErrorCorrection30Percent ErrorCorrectionLevel = 3
)

// NewEncoder returns qr encoder with the given type and error correction level.
// The encoder type determines the output format of the qr code.
// The error correction level determines the amount of data that can be recovered from the qr code.
// The encoder type must be one of the following: TextDarkMode, TextLightMode, HTMLMode
// The error correction level must be one of the following: ErrorCorrection7Percent, ErrorCorrection15Percent, ErrorCorrection25Percent, ErrorCorrection30Percent
func NewEncoder(encoderType EncoderType, errorCorrectionLevel ErrorCorrectionLevel) (*Encoder, error) {
	var q Encoder
	switch encoderType {
	case TextDarkMode:
		q.rc = &darkMode
		q.strFunc = text
		break
	case TextLightMode:
		q.rc = &lightMode
		q.strFunc = text
		break
	case HTMLMode:
		q.strFunc = html
		break
	case TerminalMode:
		q.rc = &darkMode
		q.strFunc = func(rc *runeCol, code *image.Image, headers *[]string) (string, error) {
			s, e := text(rc, code, headers)
			if e != nil {
				return "", e
			}
			front := "\033[40;97m"
			back := "\033[0m\n"
			s = strings.ReplaceAll(s, "\n", back+front)
			return front + strings.TrimSuffix(s, front), nil
		}
		break
	default:
		return nil, fmt.Errorf("invalid encoder type: %d", encoderType)
	}
	if errorCorrectionLevel < 0 || errorCorrectionLevel > 3 {
		return nil, fmt.Errorf("invalid error correction level: %d", errorCorrectionLevel)
	}
	q.errCorr = errorCorrectionLevel
	return &q, nil
}
