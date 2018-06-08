ELM_FILES = $(shell find . -path ./elm-stuff -prune -o -type f -name '*.elm')

build: build-elm build-css
	mkdir -p ../assets/js
	mkdir -p ../assets/css
	cp elm.js ../assets/js/elm.js
	cp elm-mdc.js ../assets/js/elm-mdc.js
	cp material-components-web.css ../assets/css/material-components-web.css

build-elm:
	elm-make --yes src/Main.elm --output elm.js

build-css:
	./node_modules/.bin/webpack
	rm -f style-bundle.js

setup:
	npm i
	cd node_modules/elm-mdc/ && make elm-mdc.js
	cp node_modules/elm-mdc/elm-mdc.js .
	cd node_modules/elm-mdc/ && make distclean

run:
	elm-reactor

clean:
	rm -f material-components-web.css style-bundle.js elm.js
	rm -rf elm-stuff/build-artifacts

distclean: clean
	rm -rf elm-stuff
	rm -rf elm-mdc.js
	rm -rf node_modules
