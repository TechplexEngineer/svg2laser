svg2laser
=========

## Goal
Streamline hte process of cutting FRC Robotics prototypes designed in Onshape on an Epilog Helix 18"x24" laser cutter.


This repo contains a number of tests and ideas, but nothing works great yet.

## Data flow
1. Possible sources (Suppored via Onshape export):
    - PDF
    - DWG -- obscure format
    - DXF -- obscure format
    - DWT -- obscure format
    - SVG ==> Easy to modify
    - PNG -- eliminated, not vector
    - JPEG -- eliminated, not vector


2. Set line thickness to .001"
    Math and text substitution will get us most of the way there

3. Convert SVG to Postscript
inkscape input.svg --export-filename=output.ps

4. Send postscript through epilog postprocessor
liblasercut


## Inspiration
- LibLaserCut
- Lathser
- ctrl-cut


## Notes

> The initial value for SVG user coordinates is that 1 user unit equals one CSS "px" (pixel) unit. By CSS standards, a "px" unit is exactly equal to 1/96th of an "in" (inch) unit. If you scale your SVG with transforms or a viewBox attribute, all the length units scale accordingly, so the ratio remains constant.
src: https://stackoverflow.com/a/23096315/429544

1 (svg user coordinate) = 1 (css px)
1 (css px) = 1/96 (inch) ---- or ---- 96 (css px) = 1 (inch)

if I have .3 (svg user coordinate) that is the same as .3 (css px)

So to convert .3 (css px) to inches we divide by 96
.3/96 = 0.003125 (inch)

This seems wrong because if one uses inkscape to change stroke-width to .001" the resulting file has .3 as the stroke width

So maybe 96 is wrong....

What value of inches is actually per css px

.3 (css px) / x (css px/inch) = .001 (inch)
Cross multiply (both sides by x) and divide (.001)

.3 / .001 = x = 300

How can we calculate the resolution if its not a fixed value...


From a pseudocode perspective:
Take a hypothetical file
```
<svg
	width="{width}"
	height="{height}"
	viewBox="{x_min} {y_min} {viewWidth} {viewHeight}"
```


.001" * {width}/({viewWidth}-{x_min}) = {new stroke width}

big thanks to this so post for leading me down this thought path
https://stackoverflow.com/questions/23068907/how-do-i-use-inches-with-snap-svg#:~:text=The%20initial%20value%20for%20SVG,so%20the%20ratio%20remains%20constant.