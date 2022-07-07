

1. Possible sources:
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
