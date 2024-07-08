---
# vim: set filetype=yaml:
activity:
  mtb: true
  type: mtb
  date: {{ .Date }}
  title: {{ .Title }}
  pointOfOrigin:
    name: \framebox{P} xx
    qr: {{ .StartLocationQr }}
    region: Allgäu -
  season: {{ .Season }} {{ .Year }}
  rating: ${{ range .Stars }}\bigstar~{{ end }}$
  company: {{ .Company }}
  difficulty: {{ .Difficulty }}
  restaurant: {{ .Restaurant }}

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
