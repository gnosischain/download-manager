package main

import (
	"github.com/mgutz/ansi"
)

// https://github.com/mgutz/ansi
// https://en.wikipedia.org/wiki/ANSI_escape_code#Colors

// WhiteWithBlueBackgroundColor ...
var WhiteWithBlueBackgroundColor = ansi.ColorFunc("white+h:33+h")

// WhiteWithGreenBackgroundColor ...
var WhiteWithGreenBackgroundColor = ansi.ColorFunc("white+h:green")

// WhiteBrightWithRedBackgroundColor ...
var WhiteBrightWithRedBackgroundColor = ansi.ColorFunc("white+h:red")

// WhiteBrightWithOrangeBackgroundColor ...
var WhiteBrightWithOrangeBackgroundColor = ansi.ColorFunc("white+h:202+h")

// BlueBoldBrightColor ...
var BlueBoldBrightColor = ansi.ColorFunc("27+b+h")

// GreenBoldBrightColor ...
var GreenBoldBrightColor = ansi.ColorFunc("green+b+h")

// GreenLightBrightColor ...
var GreenLightBrightColor = ansi.ColorFunc("green+h")

// CyanBoldBrightColor ...
var CyanBoldBrightColor = ansi.ColorFunc("cyan+b+h")

// YellowBoldBrightColor ...
var YellowBoldBrightColor = ansi.ColorFunc("yellow+b+h")

// RedBoldBrightColor ...
var RedBoldBrightColor = ansi.ColorFunc("red+b+h")

// OrangeBoldBrightColor ...
var OrangeBoldBrightColor = ansi.ColorFunc("202+b+h")

// WhiteBoldBrightColor ...
var WhiteBoldBrightColor = ansi.ColorFunc("white+b+h")

// WhiteBrightColor ...
var WhiteBrightColor = ansi.ColorFunc("white+h")

// PlainHeaderColor ...
var PlainHeaderColor = ansi.ColorFunc("6")

// PlainHeaderBoldBrightColor ...
var PlainHeaderBoldBrightColor = ansi.ColorFunc("6+b+h")

// MagentaBoldBrightColor ...
var MagentaBoldBrightColor = ansi.ColorFunc("magenta+b+h")

// PurpleBoldBrightColor ...
var PurpleBoldBrightColor = ansi.ColorFunc("5+b+h")

// GrayBoldBrightColor ...
var GrayBoldBrightColor = ansi.ColorFunc("8+b+h")

// PercentageColorFont ...
var PercentageColorFont = func(percent int) func(string) string {
	switch {
	case percent == 0:
		return ansi.ColorFunc("88+h")
	case 0 < percent && percent <= 10:
		return ansi.ColorFunc("124+h")
	case 10 < percent && percent <= 20:
		return ansi.ColorFunc("160+h")
	case 20 < percent && percent <= 30:
		return ansi.ColorFunc("196+h")
	case 30 < percent && percent <= 40:
		return ansi.ColorFunc("166+h")
	case 40 < percent && percent <= 50:
		return ansi.ColorFunc("202+h")
	case 50 < percent && percent <= 60:
		return ansi.ColorFunc("214+h")
	case 60 < percent && percent <= 70:
		return ansi.ColorFunc("226+h")
	case 70 < percent && percent <= 80:
		return ansi.ColorFunc("190+h")
	case 80 < percent && percent <= 90:
		return ansi.ColorFunc("154+h")
	default:
		return ansi.ColorFunc("10+h")
	}
}

// ImpactColorFont ...
var ImpactColorFont = func(percent float64) func(string) string {
	switch {
	case percent == 0:
		return ansi.ColorFunc("10+h")
	case 0 < percent && percent <= 1:
		return ansi.ColorFunc("154+h")
	case 1 < percent && percent <= 5:
		return ansi.ColorFunc("190+h")
	case 5 < percent && percent <= 10:
		return ansi.ColorFunc("214+h")
	case 10 < percent && percent <= 20:
		return ansi.ColorFunc("202+h")
	case 20 < percent && percent <= 30:
		return ansi.ColorFunc("166+h")
	case 30 < percent && percent <= 40:
		return ansi.ColorFunc("196+h")
	case percent > 40:
		return ansi.ColorFunc("124+h")
	default:
		return ansi.ColorFunc("88+h")
	}
}
