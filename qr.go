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

type qrEncoder struct {
	strFunc func(rc *runeCol, code *image.Image, headers *[]string) (string, error)
	rc      *runeCol
	errCorr errorCorrectionLevel
}

func (q *qrEncoder) Encode(data string, header ...string) (string, error) {
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
	return strFunc(q.rc, &code, &header)
}

func text(rc *runeCol, code *image.Image, headers *[]string) (string, error) {
	if rc == nil || code == nil {
		return "", fmt.Errorf("encoder misconfigured, use NewEncoder when creating it")
	}
	var output = ""
	dx := (*code).Bounds().Dx()
	dy := (*code).Bounds().Dy()
	wr := rc.getRune(color.White, color.White)
	prefix := ""
	suffix := "\n"

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
		output += string(whole) + pad(dx+2, lower) + string(whole) + "\n"
	}

	return output, nil
}

func html(rc *runeCol, code *image.Image, headers *[]string) (string, error) {
	var output = ""
	return output, nil
}

type encoderType int
type errorCorrectionLevel qr.ErrorCorrectionLevel

const (
	// TextDarkMode makes qr codes for printing on dark backgrounds with white text,
	// like a terminal or a screen with dark/night mode enabled.
	// MUST BE PRINTED/DISPLAYED USING A MONOSPACE FONT.
	TextDarkMode encoderType = 0
	// TextLightMode makes qr codes for printing on light backgrounds with black text,
	// like paper or a screen with light mode.
	// MUST BE PRINTED/DISPLAYED USING A MONOSPACE FONT.
	TextLightMode encoderType = 1
	// HTMLMode makes qr codes for embedding in HTML documents or web pages.
	// Generates a table using HTML tags, does not require a monospace font.
	HTMLMode encoderType = 2

	// ErrorCorrection7Percent indicates 7% of lost data can be recovered, makes the qr code smaller
	ErrorCorrection7Percent errorCorrectionLevel = 0
	// ErrorCorrection15Percent indicates 15% of lost data can be recovered, default
	ErrorCorrection15Percent errorCorrectionLevel = 1
	// ErrorCorrection25Percent indicates 25% of lost data can be recovered, makes the qr code bigger
	ErrorCorrection25Percent errorCorrectionLevel = 2
	// ErrorCorrection30Percent indicates 30% of lost data can be recovered, makes the qr code very big
	ErrorCorrection30Percent errorCorrectionLevel = 3
)

// NewEncoder returns a new qrEncoder code with the given data and headers
func NewEncoder(encoderType encoderType, errorCorrectionLevel errorCorrectionLevel) (*qrEncoder, error) {
	var q qrEncoder
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
	default:
		return nil, fmt.Errorf("invalid encoder type: %d", encoderType)
	}
	if errorCorrectionLevel < 0 || errorCorrectionLevel > 3 {
		return nil, fmt.Errorf("invalid error correction level: %d", errorCorrectionLevel)
	}
	q.errCorr = errorCorrectionLevel
	return &q, nil
}
