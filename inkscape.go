package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

// svgConvertBuffer uses inkscape to convert to a supported output format. (pdf default)
// By default supported output types are: svg, png, ps, eps, pdf, emf, wmf
// see: https://inkscape.org/doc/inkscape-man.html#export-type-TYPE-TYPE
func svgConvertBuffer(inputBytes []byte, outputBuffer io.Writer, stderrBuffer io.Writer) error {
	inkscapeExecutable := "inkscape"
	if exePath, isSet := os.LookupEnv("SVG2LASER_INKSCAPE_PATH"); isSet {
		inkscapeExecutable = exePath
	}
	return execPipe(inkscapeExecutable, []string{"--pipe", "--export-filename=-", "--export-type=pdf"}, inputBytes, outputBuffer, stderrBuffer)
}

func execPipe(bin string, args []string, stdin []byte, outputBuffer io.Writer, stderrBuffer io.Writer) error {

	cmd := exec.Command(bin, args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if cmd.Start() != nil {
		return err
	}

	_, err = in.Write(stdin)
	if err != nil {
		return err
	}
	if in.Close() != nil {
		return err
	}

	fileBytes, err := ioutil.ReadAll(out)
	if err != nil {
		return err
	}
	if out.Close() != nil {
		return err
	}

	_, err = io.Copy(outputBuffer, bytes.NewReader(fileBytes))
	if err != nil {
		return err
	}

	stderrB, err := ioutil.ReadAll(stderr)
	if err != nil {
		return err
	}

	if stderr.Close() != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 3 {
			log.Printf("ERROR: %s - %s\n", stderrB, err)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
