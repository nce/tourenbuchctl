# vim: set ft=gnuplot:

# this creates the default layout with axis labels on the left and continuous
# grid.
# default for: mtb, hike, ...

load "activity-settings.plt"

myScaleKM = {{ .Activity }}KM
myScaleHM = {{ .Activity }}HM
myTics = {{ .Activity }}Tics
myColor = {{ .Activity }}Color

file = 'gpxdata.txt'

reset
unset key
stats file u 3:2 nooutput
set xrange [STATS_min_x:STATS_max_x]
# floor to lowest elevation, to always display minimum elevation;
set yrange [floor(STATS_min_y/myTics) * myTics:STATS_max_y]
set y2range [floor(STATS_min_y/myTics) * myTics:STATS_max_y]

# Max Km mit Skalierungsfaktor. Den (fiktiven) Pixelwert von 300dpi auf cm umrechnen und den gesamten Skalierungsfaktor darauf mulitiplizieren.
# vielleicht nicht so ideal
w = floor(STATS_max_x * myScaleKM) * 2.54 / 300 * 2
# h = floor((STATS_max_y - STATS_min_y) * myScaleHM) * 2.54 / 300 * 2
# switch from 300 to 250 as small tours are getting dense 07.2016
h = floor((STATS_max_y - STATS_min_y) * myScaleHM) * 2.54 / 250 * 2

# -- [ Labels ] --
load "elevation.plt"

# -- [ Terminal ] --
# epslates does not support transparency
# for mac xquartz is required to install epslatex https://xquartz.macosforge.org/landing/
# brew install gnuplot --cairo
set terminal cairolatex size w cm,h cm color
set output "elevation.tex"

# -- [ Grid and Tics ] --
set xtic 5
set xtic nomirror
unset border
set grid lt 0 dashtype 2 lw 3 lc rgb "black"
set format x "%g\\tiny\\,\\color{darkgray}{km}"

# left labeled axis
set ytics border myTics nomirror
set format y "%h\\tiny\\,\\color{darkgray}{m}"
# cut right margin
set rmargin at screen 1

set border 3 lt 3 lw 3 lc rgb "#708090"

# correct the margin calculations which are based
# on the length of the format string, to a fixed value
set lmargin 4.8

set style fill transparent solid 0.45 noborder
plot file u 3:2 w filledcurve x1 lc rgb "black" , \
  file u 3:2 w lines lt 1 lw 5 lc rgb myColor
