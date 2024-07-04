---
# vim: set filetype=yaml:
activity:
  mtb: true
  type: mtb
  date: {{ .Date }}
  title: {{ .Name }}
  pointOfOrigin:
    name: \framebox{P} xx
    qr:
    region: Allgäu -
  season: Sommer {{ .Year }}
  rating: ${{ range .Stars }}\bigstar~{{ end }}$
  company:
  difficulty: S2
  restaurant:

stats:
  ascent:
  distance:
  movingTime:
  overallTime:
  startTime:
# puls:

layout:
  #headElevationProfile: true
  # default
  tableSize: 0.50
  mapSize: 0.49
  mapHeight: 17
  # adjust if excess of space
  linespread: 1.05
  elevationProfileRightMargin: 0
...
foobar

\noindent $\blacktriangleright$ barfoo

\input{\textpath/img-even}
