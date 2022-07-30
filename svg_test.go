package main

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func getFuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	fmt.Printf("%s\n", frame.Function)
	substr := ".Test"
	return frame.Function[strings.LastIndex(frame.Function, substr)+len(substr):]
}

func TestSVGAttrs_getResolutionPxPerIn(t *testing.T) {

	tests := []struct {
		name    string
		a       SVGAttrs
		want    int
		wantErr bool
	}{
		{
			name: "happy path - mm",
			a: SVGAttrs{
				width:   "457.2mm",
				height:  "457.2mm",
				viewbox: "0 0 5400 5400",
			},
			want:    300,
			wantErr: false,
		}, {
			name: "happy path - in",
			a: SVGAttrs{
				width:   "2in",
				height:  "2in",
				viewbox: "0 0 600 600",
			},
			want:    300,
			wantErr: false,
		}, {
			name: "mismatch dpi",
			a: SVGAttrs{
				width:   "2in",
				height:  "3in",
				viewbox: "0 0 5400 5400",
			},
			want:    0,
			wantErr: true,
		}, {
			name: "invalid viewbox - string",
			a: SVGAttrs{
				width:   "457.2mm",
				height:  "457.2mm",
				viewbox: "0 0 5400 5400mm",
			},
			want:    0,
			wantErr: true,
		}, {
			name: "invalid viewbox - missing number",
			a: SVGAttrs{
				width:   "457.2mm",
				height:  "457.2mm",
				viewbox: "0 0 5400",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.a.getResolutionPxPerIn()
			if (err != nil) != tt.wantErr {
				t.Errorf("getResolutionPxPerIn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getResolutionPxPerIn() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_widthHeightRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "happy path - in", input: "2in", want: []string{"2", "in"}},
		{name: "happy path - mm", input: "50.8mm", want: []string{"50.8", "mm"}},
		{name: "invalid - missing units", input: "2", want: nil},
		{name: "invalid - blank", input: "", want: nil},
	}
	funcName := getFuncName()
	funcName = strings.TrimPrefix(funcName, "_")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := widthHeightRegex.FindAllStringSubmatch(tt.input, -1) //-1 means all matches
			if tt.want == nil && matches == nil {
				return //success
			}
			if matches == nil {
				t.Errorf("%s input '%s' not matched", funcName, tt.input)
				return
			}
			got := matches[0]

			expected := append([]string{tt.input}, tt.want...)
			if !reflect.DeepEqual(expected, got) {
				t.Errorf("%s expected %v got %v", funcName, expected, got)
				return
			}
		})
	}
}

func Test_viewBoxRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "happy path", input: "0 0 5400 5400", want: []string{"0", "0", "5400", "5400"}},
		{name: "invalid - string", input: "0 0 5400 5400mm", want: nil},
		{name: "invalid - missing number", input: "0 0 5400", want: nil},
		{name: "invalid - blank", input: "", want: nil},
	}
	funcName := getFuncName()
	funcName = strings.TrimPrefix(funcName, "_")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := viewBoxRegex.FindAllStringSubmatch(tt.input, -1) //-1 means all matches
			if tt.want == nil && matches == nil {
				return //success
			}
			if matches == nil {
				t.Errorf("%s input '%s' not matched", funcName, tt.input)
				return
			}
			got := matches[0]

			expected := append([]string{tt.input}, tt.want...)
			if !reflect.DeepEqual(expected, got) {
				t.Errorf("%s expected %v got %v", funcName, expected, got)
				return
			}
		})
	}
}
