# tourenbuchctl
A tool for working with my tourenbuch.

`Tourenbuch` is a digital & printed log book of my personal outdoor sports activities.
Each entry consists of gpx tracks, map, stats, plotted elevation graphs and an
overall summary.

This CLI helps me to interact with it. Until i release more information about
my (private) Tourenbuch, it's  probabaly not useful for anyone.

## Strava
Activity stats like distance or climbed elevation gets queried from strava and
parsed in the Tourenbuch.

# Dev

## Prerequisites

Go-swagger is incompatible with 3.x api defintion of strava...

Install swagger-codegen (java)
```
brew install swagger-codegen
```
