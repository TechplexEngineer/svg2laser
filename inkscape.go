package main

import (
	"fmt"
	"github.com/magefile/mage/sh"
	"log"
	"os"
)

// svgConvert uses inkscape to convert to a supported output format.
// By default supported output types are: svg, png, ps, eps, pdf, emf, wmf
// see: https://inkscape.org/doc/inkscape-man.html#export-type-TYPE-TYPE
func svgConvert(outputFileName string, inputFileName string) error {
	inkscapeExecutable := "inkscape"
	if exePath, isSet := os.LookupEnv("SVG2LASER_INKSCAPE_PATH"); isSet {
		inkscapeExecutable = exePath
	}

	//inkscape --export-filename=filename.pdf filename.svg
	log.Printf("command: %s %s %s %s", inkscapeExecutable, "--export-filename", outputFileName, inputFileName)
	err := sh.Run(inkscapeExecutable, "--export-filename", outputFileName, inputFileName)
	if err != nil {
		return fmt.Errorf("error running inkscape - %w", err)
	}
	return nil
}
