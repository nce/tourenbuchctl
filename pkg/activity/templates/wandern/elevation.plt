# vim: set ft=gnuplot:
set label 1 sprintf("\\textcolor\{wandern\}\{%s \\tiny %sm\}", title, maxElevation) at STATS_pos_max_y,STATS_max_y  point pointtype 7 ps 0.6 offset 0.3,0.3 front
