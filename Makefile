.DEFAULT_GOAL := build
DIR ?= $(shell echo `pwd`/mocks/content/)
TAG_FILE ?= $(shell echo `pwd`/mocks/content/dir2/tags.txt)
GO_ENV ?= development

# Golang and yarn need to be on the system. They will do the rest
.PHONY: install
install:
	go get contented
	yarn install

# This should be the most common development experience (but a little awkward based on docker)
.PHONY: setup
setup:
	make db-reset
	make tags
	make db-seed
	make preview
	make typescript
	make dev

# TODO: GoBuffalo is deprecated and the docker image must be redone.
.PHONY: build
build:
	docker build -f Dockerfile -t contented:latest .

.PHONY: dev
dev:
	export DIR=$(DIR) TAG_FILE=$(TAG_FILE) && go run cmd/app/main.go

# I would rather use gotestsum, but buffalo does a bunch of DB setup that doesn't play
# nice with go test or gotestsum. Or potentially my tests need some saner / better init
# data around the environment? Should probably move preview into it's own package with how 
# damn slow ffmpeg seek screen tests are on MacOSX.
.PHONY: test
test:
	export GO_ENV=test && export DIR=$(DIR) && go run gotest.tools/gotestsum@latest --format testname ./pkg/worker
	export GO_ENV=test && export DIR=$(DIR) && go run gotest.tools/gotestsum@latest --format testname ./pkg/models
	export GO_ENV=test && export DIR=$(DIR) && go run gotest.tools/gotestsum@latest --format testname ./pkg/managers
	export GO_ENV=test && export DIR=$(DIR) && go run gotest.tools/gotestsum@latest --format testname ./pkg/actions
	export GO_ENV=test && export DIR=$(DIR) && go run gotest.tools/gotestsum@latest --format testname ./pkg/utils

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

.PHONY: db-reset
db-reset:
	export GO_ENV=$(GO_ENV) && go run ./cmd/scripts/cmdline.go --action reset

.PHONY: db-populate
db-populate:
	export GO_ENV=$(GO_ENV) && export DIR=$(DIR) && go run ./cmd/scripts/cmdline.go --action populate

.PHONY: preview
preview:
	export GO_ENV=$(GO_ENV) && export DIR=$(DIR) && go run ./cmd/scripts/cmdline.go --action preview

.PHONY: encode
encode:
	export GO_ENV=$(GO_ENV) && export DIR=$(DIR) && go run ./cmd/scripts/cmdline.go --action encode

.PHONY: find-dupes
find-dupes:
	export GO_ENV=$(GO_ENV) && export DIR=$(DIR) && go run ./cmd/scripts/cmdline.go --action duplicates

# Read from a tag file and import the tags to the DB
.PHONY: tags
tags:
	export GO_ENV=$(GO_ENV) && export DIR=$(DIR) && export TAG_FILE=$(TAG_FILE) && go run ./cmd/scripts/cmdline.go --action tags
