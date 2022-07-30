package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

// svgConvert uses inkscape to convert to a supported output format.
// By default supported output types are: svg, png, ps, eps, pdf, emf, wmf
// see: https://inkscape.org/doc/inkscape-man.html#export-type-TYPE-TYPE
//func svgConvert(outputFileName string, inputFileName string) error {
//	inkscapeExecutable := "inkscape"
//	if exePath, isSet := os.LookupEnv("SVG2LASER_INKSCAPE_PATH"); isSet {
//		inkscapeExecutable = exePath
//	}
//
//	//inkscape --export-filename=filename.pdf filename.svg
//	log.Printf("command: %s %s %s %s", inkscapeExecutable, "--export-filename", outputFileName, inputFileName)
//	err := sh.Run(inkscapeExecutable, "--export-filename", outputFileName, inputFileName)
//	if err != nil {
//		return fmt.Errorf("error running inkscape - %w", err)
//	}
//	return nil
//}

func svgConvertBuffer(inputBytes []byte, outputBuffer io.Writer, stderrBuffer io.Writer) error {
	inkscapeExecutable := "inkscape"
	if exePath, isSet := os.LookupEnv("SVG2LASER_INKSCAPE_PATH"); isSet {
		inkscapeExecutable = exePath
	}
	return origPipe(inkscapeExecutable, []string{"--pipe", "--export-filename=-", "--export-type=pdf"}, inputBytes, outputBuffer, stderrBuffer)
}

//func check(e error) {
//	if e != nil {
//		panic(e)
//	}
//}

func origPipe(bin string, args []string, stdin []byte, outputBuffer io.Writer, stderrBuffer io.Writer) error {

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

func brokenPipe(bin string, args []string, inputBytes []byte, outputBuffer io.Writer, stderrBuffer io.Writer) error {
	cmd := exec.Command(bin, args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer in.Close()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()
	err = cmd.Start()
	if err != nil {
		return err
	}

	_, err = in.Write(inputBytes)
	if err != nil {
		return err
	}

	fileBytes, err := ioutil.ReadAll(stdout)
	_ = fileBytes
	_, err = io.Copy(outputBuffer, stdout)
	if err != nil {
		return err
	}
	if stderrBuffer != nil {
		_, err = io.Copy(stderrBuffer, stderr)
		if err != nil {
			return err
		}
	}

	if err = cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 3 {
			if err != nil {
				return err
			}
		}
	}

	return nil
}
