package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

// SVGAttrs provides processing utilities for svgs
// Where width and height have a unit of "em" | "ex" | "px" | "in" | "cm" | "mm" | "pt" | "pc" | "%"
// per https://developer.mozilla.org/en-US/docs/Web/SVG/Content_type#length
type SVGAttrs struct {
	width   string //eg "457.2mm"
	height  string //eg "457.2mm"
	viewbox string //eg "0 0 5400 5400"
}

var viewBoxRegex = regexp.MustCompile(`^(\d+) (\d+) (\d+) (\d+)$`)
var widthHeightRegex = regexp.MustCompile(`^([\d.]+)(mm|in)$`)

// should only be used when we are SURE the input string is a valid number.
// eg. when already checked via regex
func mustParseInt(input string) int {
	i, err := strconv.ParseInt(input, 10, 0) //bitSize=0 means int
	if err != nil {
		panic(err) // this should not happen if function is used properly
	}
	return int(i)
}

func mustParseFloat(input string) float64 {
	i, err := strconv.ParseFloat(input, 64) //bitSize=64 means float64
	if err != nil {
		panic(err) // this should not happen if function is used properly
	}
	return i
}

const MILIMETERS_PER_INCH = 25.4
const float64EqualityThreshold = 1e-3

func (a SVGAttrs) getResolutionPxPerIn() (int, error) {

	var widthPx int
	var heightPx int
	{
		matches := viewBoxRegex.FindAllStringSubmatch(a.viewbox, -1) //-1 means all, no limit

		if matches == nil || len(matches[0])-1 != 4 { //-1 because matches[0][0] is the full matched string
			return 0, fmt.Errorf("invalid viewbox '%s'", a.viewbox)
		}
		parts := matches[0]
		xMin, yMin, xMax, yMax := mustParseInt(parts[1]), mustParseInt(parts[2]), mustParseInt(parts[3]), mustParseInt(parts[4])
		widthPx = xMax - xMin
		heightPx = yMax - yMin
	}

	var widthIn float64
	{
		widthMatches := widthHeightRegex.FindAllStringSubmatch(a.width, -1) //-1 means all, no limit
		if widthMatches == nil || len(widthMatches[0])-1 != 2 {             //-1 because matches[0][0] is the full matched string
			return 0, fmt.Errorf("invalid width '%s'", a.width)
		}
		widthIn = mustParseFloat(widthMatches[0][1])
		widthUnit := widthMatches[0][2]
		if widthUnit != "in" && widthUnit != "mm" {
			return 0, fmt.Errorf("invalid width unit '%s'", widthUnit)
		}
		if widthUnit == "mm" {
			widthIn /= MILIMETERS_PER_INCH //convert to inches
		}
	}

	var heightIn float64
	{
		heightMatches := widthHeightRegex.FindAllStringSubmatch(a.height, -1) //-1 means all, no limit
		if heightMatches == nil || len(heightMatches[0])-1 != 2 {             //-1 because matches[0][0] is the full matched string
			return 0, fmt.Errorf("invalid height '%s'", a.height)
		}
		heightIn = mustParseFloat(heightMatches[0][1])
		heightUnit := heightMatches[0][2]
		if heightUnit != "in" && heightUnit != "mm" {
			return 0, fmt.Errorf("invalid width unit '%s'", heightUnit)
		}
		if heightUnit == "mm" {
			heightIn /= MILIMETERS_PER_INCH //convert to inches
		}
	}
	widthPxPerInch := float64(widthPx) / widthIn
	heightPxPerInch := float64(heightPx) / heightIn
	if math.Abs(widthPxPerInch-heightPxPerInch) > float64EqualityThreshold {
		return 0, fmt.Errorf("width and height pixels per inch do not match. width: %f height %f, %.15f", widthPxPerInch, heightPxPerInch, math.Abs(widthPxPerInch-heightPxPerInch))
	}

	return int(widthPxPerInch), nil
}
