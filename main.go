package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

import (
	"aqwari.net/xml/xmltree"
	"github.com/gorilla/mux"
)

//go:embed index.html
var indexTemplate string

func main() {
	serve := flag.Bool("serve", false, "enable rest api and web server. If specified f and o flags are ignored")
	port := flag.Int("port", 8080, "port to listen on. Only used if serve flag is passed")
	inFile := flag.String("f", "", "input svg file to convert to pdf for laser")
	outFile := flag.String("o", "", "output filename, defaults to input file name with -for-laser.svg appended")
	flag.Parse()

	if *serve {
		r := mux.NewRouter()
		r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintf(writer, indexTemplate)
		}).Methods(http.MethodGet)
		r.HandleFunc("/upload", func(w http.ResponseWriter, request *http.Request) {
			err := request.ParseMultipartForm(5 * 1024 * 1024) //5MB = 5 * 1024 * 1024
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest) // too big or not a multipart form upload
			}

			// The argument to FormFile must match the name attribute
			// of the file input on the frontend
			file, fileHeader, err := request.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer file.Close()

			requestId := time.Now().UnixNano()

			uploadDir := path.Join("./uploads", fmt.Sprintf("%d", requestId))

			// Create the uploads folder if it doesn't
			// already exist
			err = os.MkdirAll(uploadDir, os.ModePerm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			inFilePath := path.Join(uploadDir, filepath.Base(fileHeader.Filename))

			// Create a new file in the uploads directory
			dst, err := os.Create(inFilePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			outStream := bytes.Buffer{}
			err = fixStoke(file, &outStream, .001)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			// Copy the uploaded file to the filesystem
			// at the specified destination
			_, err = io.Copy(dst, &outStream)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fileWithoutSuffix := strings.TrimSuffix(filepath.Base(fileHeader.Filename), filepath.Ext(fileHeader.Filename))
			outFilePath := path.Join(uploadDir, fileWithoutSuffix+".pdf")
			log.Printf("converting '%s' to '%s'", inFilePath, outFilePath)
			err = svgConvert(outFilePath, inFilePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			readFile, err := ioutil.ReadFile(outFilePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Add("Content-Type", "application/pdf")
			_, _ = w.Write(readFile)
			//@todo don't use the filesystem

		}).Methods(http.MethodPost)

		err := http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
		if err != nil {
			log.Printf("Error: %s", err)
			os.Exit(1)
		}
		return
	}

	if err := fixFile(*inFile, *outFile); err != nil {
		log.Printf("Error: %s", err)
		os.Exit(1)
	}
}
func fixFile(inFile string, outFile string) error {
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
