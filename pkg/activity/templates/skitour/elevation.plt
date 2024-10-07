# vim: set ft=gnuplot:

set label 1 sprintf("\\textcolor\{skitour\}\{%s \\tiny %sm\}", ARG1, ARG2) at STATS_pos_max_y,STATS_max_y  point pointtype 7 ps 0.6 offset 0.3,0.3 front
# set label 1 '\textcolor{skitour}{ \scriptsize m}' at STATS_pos_max_y,STATS_max_y point pointtype 7 ps 0.6 offset 0.3,0.3 front
#set label 2 '\textcolor{skitour}{\tiny m}' at 3.2,1830 point pointtype 7 ps 0.6 offset -0.3,0.3 front rotate by 90
