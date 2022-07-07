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