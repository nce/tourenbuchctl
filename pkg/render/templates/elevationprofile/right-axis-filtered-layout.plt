# vim: set ft=gnuplot:

# this creates the filtered layout with axis labels on the right and a white
# overlay for everything left of the peak
# default for: skitour

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

# Max Km mit Skalierungsfaktor. Den (fiktiven) Pixelwert von 300dpi auf cm
# umrechnen und den gesamten Skalierungsfaktor darauf mulitiplizieren.
# vielleicht nicht so ideal
w = floor(STATS_max_x * myScaleKM) * 2.54 / 300	* 2
h = floor((STATS_max_y - STATS_min_y) * myScaleHM) * 2.54 / 300 * 2

# -- [ Terminal ] --
# epslates does not support transparency
# for mac xquartz is required to install epslatex https://xquartz.macosforge.org/landing/
# brew install gnuplot --cairo
set terminal cairolatex size w cm,h cm color
set output "elevation.tex"
set multiplot

# -- [ Grid and Tics ] --
set xtics 5
unset border
unset xtics
unset ytics
unset x2tics
set y2tics border myTics
set format x "%g\\tiny\\,\\color{darkgray}{km}"
set format y2 "%h\\tiny\\,\\color{darkgray}{m}"
set border 8 lt 0 lw 3 lc rgb "black"
#set ytic scale 0
set xtics
set grid y2 lt 0 lw 3 lc rgb "black"

# huge change in scaling; but removes right margin
set rmargin at screen 0.915;

# only draw values until the highest point; this is used to overwrite the leftside grid
filter(x,max) = (x <= max) ? x: 1/0

set style fill transparent solid 0.45 noborder
plot file u 3:2 w filledcurve x1 lc rgb "black" , \
  file u 3:2 w lines lt 1 lw 5 lc rgb myColor
unset grid
set style fill solid noborder
set noborder
set nolabel

# -- [ Labels ] --
load "elevation.plt"

plot file u (filter($3,STATS_pos_max_y)-0.01):($2 + 5) w filledcurve above x2 lc rgb "white"
