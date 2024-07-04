# Makefile for Tourenbuch
.PHONY: register description.tex elevation.tex

SEASON=sommer{{ .Year }}
PROJECTROOT={{ .TextLibraryPath}}
# elevation profile
GPXPLOT=meta/gpxplot.py

# get the last two directories
CWD=$(shell basename "$(shell dirname $(shell pwd))")/$(shell basename "$(shell pwd)")
# multidaytrips
#CWD=mtb/multidaytrip/$(shell basename "$(shell dirname $(shell pwd))")/$(shell basename "$(shell pwd)")
ASSETDIR={{ .AssetLibraryPath }}

# pandoc options
PANDOCFLAGS=--from markdown+tex_math_dollars \
			--variable=assetdir:$(ASSETDIR)/$(CWD) \
			--variable=path:$(CWD) \
			--template=$(PROJECTROOT)meta/tourenbuch.template

all: tourenbuch

gpxdata.txt:
	python3 $(PROJECTROOT)$(GPXPLOT) $(ASSETDIR)/$(CWD)/input.gpx > $@

elevation.tex: gpxdata.txt
	gnuplot elevation.plt

description.tex:
	pandoc $(PANDOCFLAGS) description.md -o $@

register:
# variable assignment is fine; but in ifeq not?!
# $(eval REGISTERED:=$(shell grep -q $(CWD)/description $(PROJECTROOT)meta/$(SEASON).tex; echo $$?))
# use this instead
ifeq ($(shell grep -q $(CWD)/description $(PROJECTROOT)meta/$(SEASON).tex; echo $$?),1)
	$(shell echo \\include{$(CWD)/description} >> $(PROJECTROOT)meta/$(SEASON).tex)
	@echo "$(CWD) registered in $(SEASON).tex"
else
	@echo "$(CWD) already registered in $(SEASON).tex"
endif

tourenbuch: description.tex elevation.tex register
	cd $(PROJECTROOT); \
	pdflatex tourenbuch.tex -output-directory $(PROJECTROOT)
