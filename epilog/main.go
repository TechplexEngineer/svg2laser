package main

import (
	"fmt"
	"strings"
)

const (
	SEP                     = ";"
	PJL_HEADER              = "\u001b%%-12345X@PJL JOB NAME=%s\r\n\u001bE@PJL ENTER LANGUAGE=PCL \r\n"
	PJL_FOOTER              = "\u001b%-12345X@PJL EOJ \r\n"
	PCL_COLOR_COMPONENT_ONE = "\u001b*v%dA"
	PCL_MYSTERY1            = "\u001b&y130001300003220S"
	PCL_DATESTAMP           = "\u001b&y20150311204531D"
	PCL_MYSTERY2            = "\u001b&y0V\u001b&y0L\u001b&y0T\u001b&y0C\u001b&y0Z"
	PCL_MYSTERY3            = "\u001b&z%dC"
	PCL_MYSTERY4            = "\u001b&y%dR"
	PCL_AUTOFOCUS           = "\u001b&y%dA"
	PCL_OFF_X               = "\u001b&l%dU"
	PCL_OFF_Y               = "\u001b&l%dZ"
	PCL_UPPERLEFT_X         = "\u001b&l%dW"
	PCL_UPPERLEFT_Y         = "\u001b&l%dV"
	PCL_PRINT_RESOLUTION    = "\u001b&u%dD"
	PCL_RESOLUTION          = "\u001b*t%dR"
	PCL_CENTER_ENGRAVE      = "\u001b&y%dZ"
	PCL_GLOBAL_AIR_ASSIST   = "\u001b&y%dC"
	PCL_RASTER_AIR_ASSIST   = "\u001b&z%dA"
	PCL_POS_X               = "\u001b*p%dX"
	PCL_POS_Y               = "\u001b*p%dY"
	HPGL_START              = "\u001b%1B"
	PCL_RESET               = "\u001bE"
	R_ORIENTATION           = "\u001b*r%dF"
	R_POWER                 = "\u001b&y%dP"
	R_SPEED                 = "\u001b&z%dS"
	R_BED_HEIGHT            = "\u001b*r%dT"
	R_BED_WIDTH             = "\u001b*r%dS"
	R_COMPRESSION           = "\u001b*b%dM"
	R_DIRECTION             = "\u001b&y%dO"
	R_START                 = "\u001b*r1A"
	R_END                   = "\u001b*rC"
	R_ROW_UNPACKED_BYTES    = "\u001b*b%dA"
	R_ROW_PACKED_BYTES      = "\u001b*b%dW"
	V_INIT                  = "IN"

	V_FREQUENCY = "XR%04d"

	V_POWER        = "YP%03d"
	V_SPEED        = "ZS%03d"
	V_UNKNOWN1     = "XS0"
	V_UNKNOWN2     = "XP1"
	HPGL_LINE_TYPE = "LT"
	HPGL_PEN_UP    = "PU"
	HPGL_PEN_DOWN  = "PD"
	HPGL_END       = "\u001b%0B"
)

type Cut struct {
}

func generate_prn(
	resolution int,
	enableEngraving bool,
	enableCut bool,
	centerEngrave bool,
	airAssist bool,
	width int,
	height int,
	title string,
	cuts []Cut,
) {
	raster_power := 50
	raster_speed := 50
	fmt.Printf(PJL_HEADER, title)

	fmt.Printf(PCL_AUTOFOCUS, -1)
	fmt.Printf(PCL_GLOBAL_AIR_ASSIST, func() int {
		if airAssist {
			return 1
		}
		return 0
	}())
	fmt.Printf(PCL_CENTER_ENGRAVE, func() int {
		if centerEngrave {
			return 1
		}
		return 0
	}())

	fmt.Printf(PCL_OFF_X, 0)
	fmt.Printf(PCL_OFF_Y, 0)
	fmt.Printf(PCL_PRINT_RESOLUTION, resolution)
	fmt.Printf(PCL_POS_X, 0)
	fmt.Printf(PCL_POS_Y, 0)
	fmt.Printf(PCL_RESOLUTION, resolution)
	fmt.Printf(R_ORIENTATION, 0)

	fmt.Printf(R_POWER, raster_power)
	fmt.Printf(R_SPEED, raster_speed)

	fmt.Printf(PCL_RASTER_AIR_ASSIST, func() int {
		if airAssist {
			return 2
		}
		return 0
	}())

	bed_width := 24 * int(resolution)
	bed_height := 18 * int(resolution)

	fmt.Printf(R_BED_HEIGHT, bed_height)
	fmt.Printf(R_BED_WIDTH, bed_width)
	fmt.Printf(R_COMPRESSION, 2)
	if enableCut {
		fmt.Print(HPGL_START)
		fmt.Print(V_INIT)
		fmt.Print(SEP)
		for _, cut := range cuts {
			generate_cut(cut)
		}
		fmt.Print(HPGL_END)
	}
	fmt.Print(HPGL_START)
	fmt.Print(HPGL_PEN_UP)
	fmt.Print(PCL_RESET)
	fmt.Print(PJL_FOOTER)

	fmt.Print(strings.Repeat(" ", 4092))
	fmt.Print("Mini]\n")

}

func generate_cut(cut interface{}) {
}

func main() {

}
