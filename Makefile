.DEFAULT_GOAL := build

# TODO: Flesh out more of the makefile 
# DB reset and create
# Make previews
# ENV variable for the directory to use
DIR ?= $(shell echo `pwd`/mocks/content/)
TAG_FILE ?= $(shell echo `pwd`/mocks/content/dir2/tags.txt)

# You are going to need to have buffalo installed https://gobuffalo.io/documentation/getting_started/installation/
.PHONY: install
install:
	go get contented
	buffalo plugins install
	yarn install

# Typically you will want to have created the db and ensured it is configured (make db-create)
# And have the docker or postgres instance configured.  This take will reset the DB, import
# your tags and then populate the db, finally it creates previews.  Doing this in order
# will get the actual DB done.  When not in the DB just running make preview is typically enough.
# Finally it starts the dev server so you can view content at http://localhost:3000
.PHONY: setup
setup:
	make db-reset
	make tags
	make db-seed
	make preview
	make dev

# Need to fix the docker build, it is pretty old.
.PHONY: build
build:
	docker build .

.PHONY: dev
dev:
	export DIR=$(DIR) && TAG_FILE=$(TAG_FILE) && buffalo dev

# I would rather use gotestsum, but buffalo does a bunch of DB setup that doesn't play
# nice with go test or gotestsum. Or potentially my tests need some saner / better init
# data around the environment? Should probably move preview into it's own package with how 
# damn slow ffmpeg seek screen tests are on MacOSX.
.PHONY: test
test:
	export DIR=$(DIR) && buffalo test ./models ./utils ./managers ./actions

# This works with gotestsum, something about a DB reset is missing or magical Buffalo code.
# The Database side of things doesn't get created with gotestsum yet
# To run one test with gotestsum you can steal this line and pass --run <TestName>
.PHONY: gotestsum
gtest:
	export DIR=$(DIR) && gotestsum --format testname ./models
	export DIR=$(DIR) && gotestsum --format testname ./utils
	export DIR=$(DIR) && buffalo test ./managers
	export DIR=$(DIR) && buffalo test ./actions

.PHONY: ngdev
ngdev:
	yarn run ng build contented --configuration=dev --watch=true --base-href /public/build/

.PHONY: ngtest
ngtest:
	yarn run ng test

# Often a run with eslint --fix will actually handle just about everything
.PHONY: lint
lint:
	yarn run lint

# Typically you want a different window doing your jsbuilds and golang stuff for sanity
.PHONY: typescript
typescript:
	yarn run ng build contented --configuration=production --watch=false --base-href /public/build/

.PHONY: db-create
db-create:
	buffalo db create

.PHONY: reset-db
db-reset:
	buffalo db migrate
	buffalo db reset

.PHONY: db-populate
db-seed:
	export DIR=$(DIR) && buffalo task db:seed

.PHONY: preview
preview:
	export DIR=$(DIR) && buffalo task db:preview

.PHONY: encode
encode:
	export DIR=$(DIR) && buffalo task db:encode

# Read from a tag file and import the tags to the DB
.PHONY: tags
tags:
	echo "Looking for tagfile a" $TAG_FILE
	export TAG_FILE=$(TAG_FILE) && buffalo task db:tags

