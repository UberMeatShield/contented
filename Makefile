.DEFAULT_GOAL := build

.PHONY: build
build:
	docker build .

# I would rather use gotestsum, but buffalo does a bunch of DB setup that doesn't play
# nice with go test or gotestsum. Or potentially my tests need some saner / better init
# data around the environment? Should probably move preview into it's own package with how 
# damn slow ffmpeg seek screen tests are on MacOSX.
.PHONY: test
test:
	export DIR=`pwd`/mocks/content && buffalo test ./models ./utils ./managers ./actions

.PHONY: dev
dev:
	export DIR=`pwd`/mocks/content && buffalo dev

.PHONY: install
install:
	go get contented
	yarn install

# Typically you want a different window doing your jsbuilds and golang stuff for sanity
.PHONY: jsdev
jsdev:
	yarn run gulp typescript

# Angular is complaining about deploy urls but not using it doesn't work as well with a go dev server
.PHONY: jsprod
jsprod:
	yarn run gulp typescriptProd

# Often a run with eslint --fix will actually handle just about everything
.PHONY: lint
jslint:
	yarn run lint

# Running the basic lint
.PHONY: jsci
jsci:
	yarn run typescriptTests
	yarn run lint


