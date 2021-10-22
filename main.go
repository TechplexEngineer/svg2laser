package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
)

func run(inFile string) error {
	file, err := os.Open(inFile)
	if err != nil {
		return fmt.Errorf("unable to open %s - %w", inFile, err)
	}
	defer file.Close()

	outStream := bytes.Buffer{}

	if err := fixStoke(file, &outStream); err != nil {
		return fmt.Errorf("unable to fixStroke - %w", err)
	}

	fmt.Printf("%s", outStream.String())

	return nil
}

func main() {
	file := "./samples/drill_drawer_2/Drill_Drawer_Drawings_2.svg"

	if err := run(file); err != nil {
		log.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
}

func fixStoke(inStream io.Reader, outStream io.Writer) error {
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
