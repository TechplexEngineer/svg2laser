package main

import (
	"aqwari.net/xml/xmltree"
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

func run(inFile string, outFile string) error {
	file, err := os.Open(inFile)
	if err != nil {
		return fmt.Errorf("unable to open %s - %w", inFile, err)
	}
	defer file.Close()

	outStream := bytes.Buffer{}

	desiredStrokeWidthIn := .001

	if err := fixStoke(file, &outStream, desiredStrokeWidthIn); err != nil {
		return fmt.Errorf("unable to fixStroke - %w", err)
	}

	if len(outFile) == 0 {
		outFile = strings.TrimSuffix(inFile, ".svg") + "-for-laser.svg"
	}

	err = ioutil.WriteFile(outFile, outStream.Bytes(), fs.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	inFile := flag.String("f", "", "input svg file to convert to pdf for laser")
	outFile := flag.String("o", "", "output filename, defaults to input file name with -for-laser.svg appended")
	flag.Parse()

	if err := run(*inFile, *outFile); err != nil {
		log.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
}

func fixStoke(inStream io.Reader, outStream io.Writer, desiredStrokeWidthIn float64) error {
	file, err := ioutil.ReadAll(inStream)
	if err != nil {
		return fmt.Errorf("unable to read file to fix strokes - %w", err)
	}
	rootEle, err := xmltree.Parse(file)
	if err != nil {
		return fmt.Errorf("unable to parse file to fix strokes - %w", err)
	}

	if rootEle.StartElement.Name.Local != "svg" {
		return fmt.Errorf("root element is not svg, it is '%s'", rootEle.StartElement.Name.Local)
	}
	attrs := SVGAttrs{
		width:   rootEle.Attr("", "width"),
		height:  rootEle.Attr("", "height"),
		viewbox: rootEle.Attr("", "viewBox"),
	}

	resPxPerIn, err := attrs.getResolutionPxPerIn()
	if err != nil {
		return fmt.Errorf("fixStoke - %w", err)
	}

	desiredStrokeWidthSVGUnits := desiredStrokeWidthIn * float64(resPxPerIn)

	// find elements with a stroke-width that need to be changed
	elementsNeedChanges := rootEle.SearchFunc(func(ele *xmltree.Element) bool {
		for _, attr := range ele.StartElement.Attr {
			if attr.Name.Local == "stroke-width" {
				return true
			}
		}
		return false
	})
	// change the stroke width
	for _, ele := range elementsNeedChanges {
		ele.SetAttr("", "stroke-width", fmt.Sprintf("%.3f", desiredStrokeWidthSVGUnits))
	}

	// output the resulting file
	_, err = fmt.Fprintf(outStream, xml.Header)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(outStream, "%s", rootEle)
	if err != nil {
		return err
	}
	return nil
}

// this works but it assumes that the documents will always be 300 pixels per inch
func fixStrokeAssume300PxPerInch(inStream io.Reader, outStream io.Writer) error {
	re := regexp.MustCompile(`stroke-width="([\d.]+)"`)

	scanner := bufio.NewScanner(inStream)
	for scanner.Scan() {
		line := re.ReplaceAllString(scanner.Text(), `stroke-width=".3"`)
		_, err := outStream.Write([]byte(line + "\n"))
		if err != nil {
			return fmt.Errorf("error writing to outStream - %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading infile - %w", err)
	}

	return nil
}
