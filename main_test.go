package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

import (
	"aqwari.net/xml/xmltree"
	"github.com/matryer/is"
)

// SVG is a SVG document
type SVG struct {
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
	Doc    string `xml:",innerxml"`
}

// Store unparsed tags in a map (useful with flat dynamic xml).
// Inspired from: https://stackoverflow.com/questions/30928770/marshall-map-to-xml-in-go/33110881
// The main difference is that it can be mixed with defined tags in a struct.
//For each unparsed element UnmarshalXML is called once.

// NOTE: to be used in flat xml part with distinct tag names
// If it's not flat: the hash will contains key with empty values
// If there is several tags with the same name : only the last value will be stored

// UnparsedTag contains the tag informations
type UnparsedTag struct {
	XMLName xml.Name
	Content string `xml:",chardata"`
	//FullContent   string `xml:",innerxml"` // for debug purpose, allow to see what's inside some tags
}

// UnparsedTags store tags not handled by Unmarshal in a map, it should be labelled with `xml",any"`
type UnparsedTags map[string]string

func (m *UnparsedTags) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if *m == nil {
		*m = UnparsedTags{}
	}

	e := UnparsedTag{}
	err := d.DecodeElement(&e, &start)
	if err != nil {
		return err
	}

	//if _, ok := (*m)[e.XMLName.Local]; ok {
	//	return fmt.Errorf("UnparsedTags: UnmarshalXML: Tag %s:  multiple entries with the same name", e.XMLName.Local)
	//}
	(*m)[e.XMLName.Local] = e.Content

	return nil
}

func (u *UnparsedTags) GetContentByName(name string) string {
	return ((map[string]string)(*u))[name]
}

func Test_ted(t *testing.T) {
	is := is.New(t)

	f, err := os.Open("./samples/drill_drawer_2/Drill_Drawer_Drawings_2.svg")
	is.NoErr(err)
	defer f.Close()

	elements := make([]UnparsedTag, 0)

	if err := xml.NewDecoder(f).Decode(&elements); err != nil {
		is.NoErr(err)
	}

	log.Printf("%#v", elements)
}

func Test_xmltree(t *testing.T) {
	is := is.New(t)

	file, err := ioutil.ReadFile("./samples/drill_drawer_2/Drill_Drawer_Drawings_2.svg")
	is.NoErr(err)

	rootEle, err := xmltree.Parse(file)
	is.NoErr(err)

	//indent, err := json.MarshalIndent(tree, "", "    ")
	//is.NoErr(err)

	//log.Printf("tree: %s", indent) // this might work!!! Ignore content

	DFS(rootEle, func(ele *xmltree.Element) {

	})
}

func Test_DFS(t *testing.T) {
	is := is.New(t)

	data := `
<first>
	<second_1>
		<third_1_1></third_1_1>
		<third_1_2></third_1_2>
		<third_1_3></third_1_3>
	</second_1>
	<second_2>
		<third></third>
	</second_2>
</first>`

	root, err := xmltree.Parse([]byte(data))
	is.NoErr(err)

	visitedElements := make([]*xmltree.Element, 0)

	DFS(root, func(ele *xmltree.Element) {
		visitedElements = append(visitedElements, ele)
	})

	expectedNames := []string{
		"first",
		"second_2",
		"third_1_3",
		"third_1_3",
		"third_1_3",
		"second_2",
		"third",
	}

	is.Equal(len(visitedElements), len(expectedNames))

	for idx, name := range expectedNames {
		is.Equal(visitedElements[idx].Name.Local, name)
	}
}

func DFS(rootElement *xmltree.Element, visit func(ele *xmltree.Element)) {
	visit(rootElement)
	if len(rootElement.Children) == 0 {
		return // no more work to do
	}
	for _, ele := range rootElement.Children {
		DFS(&ele, visit)
	}
}

func Test_fixStoke(t *testing.T) {
	//type args struct {
	//	inStream io.Reader
	//}
	//tests := []struct {
	//	name          string
	//	args          args
	//	wantOutStream string
	//	wantErr       bool
	//}{
	//	// TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		outStream := &bytes.Buffer{}
	//		err := fixStoke(tt.args.inStream, outStream)
	//		if (err != nil) != tt.wantErr {
	//			t.Errorf("fixStoke() error = %v, wantErr %v", err, tt.wantErr)
	//			return
	//		}
	//		if gotOutStream := outStream.String(); gotOutStream != tt.wantOutStream {
	//			t.Errorf("fixStoke() gotOutStream = %v, want %v", gotOutStream, tt.wantOutStream)
	//		}
	//	})
	//}
}
