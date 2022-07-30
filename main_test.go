package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
)

import (
	"aqwari.net/xml/xmltree"
	"github.com/matryer/is"
)

func Test_xmltree(t *testing.T) {
	is := is.New(t)

	file, err := ioutil.ReadFile("./samples/drill_drawer_2/Drill_Drawer_Drawings_2.svg")
	is.NoErr(err)

	rootEle, err := xmltree.Parse(file)
	is.NoErr(err)

	log.Printf("Scope %s", rootEle.Scope)

	toJson := func(ele interface{}) string {
		indent, err := json.MarshalIndent(ele, "", "    ")
		is.NoErr(err)
		return string(indent)
	}
	_ = toJson

	if rootEle.StartElement.Name.Local != "svg" {
		return // fmt.Errorf("root element is not svg")
	}
	attrs := SVGAttrs{
		width:   rootEle.Attr("", "width"),
		height:  rootEle.Attr("", "height"),
		viewbox: rootEle.Attr("", "viewBox"),
	}

	resPxPerIn, err := attrs.getResolutionPxPerIn()
	is.NoErr(err)

	desiredStrokeWidthIn := .001
	desiredStrokeWidthSVGUnits := desiredStrokeWidthIn * float64(resPxPerIn)

	elementsNeedChanges := rootEle.SearchFunc(func(ele *xmltree.Element) bool {
		for _, attr := range ele.StartElement.Attr {
			if attr.Name.Local == "stroke-width" {
				return true
			}
		}
		return false
	})
	for _, ele := range elementsNeedChanges {
		ele.SetAttr("", "stroke-width", fmt.Sprintf("%.3f", desiredStrokeWidthSVGUnits))
	}

	log.Printf("%s", rootEle)
}
