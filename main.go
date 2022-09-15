package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

import (
	"aqwari.net/xml/xmltree"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

//go:embed index.html
var indexTemplate string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error unable to load .env file - %s", err)
		os.Exit(1)
	}

	serve := flag.Bool("serve", false, "enable rest api and web server. If specified f and o flags are ignored")
	port := flag.Int("port", 8080, "port to listen on. Only used if serve flag is passed")
	inFile := flag.String("f", "", "input svg file to convert to pdf for laser")
	outFile := flag.String("o", "", "output filename, defaults to input file name with -for-laser.svg appended")
	flag.Parse()

	if *serve {
		r := mux.NewRouter()

		tmpl, err := template.New("").Parse(indexTemplate)
		if err != nil {
			log.Printf("Error parsing index template - %s", err)
			os.Exit(1)
		}

		clientId := os.Getenv("CLIENT_ID")
		if len(clientId) <= 0 {
			log.Printf("Error CLIENT_ID env var is not set")
			os.Exit(1)
		}
		redirectUri := os.Getenv("REDIRECT_URI")
		if len(redirectUri) <= 0 {
			log.Printf("Error REDIRECT_URI env var is not set")
			os.Exit(1)
		}

		r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

			stateObj := map[string]string{
				"did":   request.FormValue("did"),
				"wvm":   request.FormValue("wvm"),
				"wvmid": request.FormValue("wvmid"),
				"eid":   request.FormValue("eid"),
			}
			stateStr, err := json.Marshal(stateObj)
			if err != nil {
				log.Printf("Error marshaling state object - %s", err)
				os.Exit(1)
			}

			//https://onshape-public.github.io/docs/oauth/
			onshapeAuthURL, err := url.Parse("https://oauth.onshape.com/oauth/authorize?response_type=code")
			if err != nil {
				log.Printf("Error unable to parse url - %s", err)
				os.Exit(1)
			}
			q := onshapeAuthURL.Query()
			q.Set("client_id", clientId)
			q.Set("redirect_uri", redirectUri)
			_ = stateStr
			//q.Set("state", string(stateStr))

			// include the company id so user's don't have to select if they are part of multiple companies
			if len(request.FormValue("company_id")) > 0 {
				q.Set("company_id", request.FormValue("company_id"))
			}
			onshapeAuthURL.RawQuery = q.Encode()

			data := map[string]any{
				"isOnshape":      len(request.FormValue("did")) > 0,
				"OnshapeAuthURL": onshapeAuthURL,
			}
			if err = tmpl.Execute(writer, data); err != nil {
				log.Printf("Error executing index template - %s", err)
			}

		}).Methods(http.MethodGet)

		r.HandleFunc("/upload", func(w http.ResponseWriter, request *http.Request) {
			err := request.ParseMultipartForm(5 * 1024 * 1024) //5MB = 5 * 1024 * 1024
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest) // too big or not a multipart form upload
			}

			// The argument to FormFile must match the name attribute
			// of the file input on the frontend
			uploadedFile, fileHeader, err := request.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer uploadedFile.Close()

			svgReadyForCutting := bytes.Buffer{}
			err = fixStoke(uploadedFile, &svgReadyForCutting, .001)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fileWithoutSuffix := strings.TrimSuffix(filepath.Base(fileHeader.Filename), filepath.Ext(fileHeader.Filename))

			pdfReadyForCutting := bytes.Buffer{}
			err = svgConvertBuffer(svgReadyForCutting.Bytes(), &pdfReadyForCutting, os.Stderr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Add("Content-Type", "application/pdf")
			action := "attachment" //attachment means download immediately
			if request.FormValue("action") == "preview" {
				action = "inline" //inline means show in page
			}

			w.Header().Add("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, action, fileWithoutSuffix+".pdf"))
			_, _ = w.Write(pdfReadyForCutting.Bytes())

		}).Methods(http.MethodPost)

		r.HandleFunc("/oauthredirect", func(writer http.ResponseWriter, request *http.Request) {

			if len(request.FormValue("error_code")) > 0 {
				u, _ := url.Parse(redirectUri) //@todo should add in the original src parameters
				http.Redirect(writer, request, u.Host, http.StatusSeeOther)
				return
			}

			code := request.FormValue("code")
			_ = code //@todo https://onshape-public.github.io/docs/oauth/#exchanging-the-code-for-a-token

			data := map[string]any{
				"isOnshape":      true,
				"OnshapeAuthURL": "",
			}
			if err = tmpl.Execute(writer, data); err != nil {
				log.Printf("Error executing index template - %s", err)
			}

		}).Methods(http.MethodGet)

		err = http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
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
