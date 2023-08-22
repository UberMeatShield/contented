.DEFAULT_GOAL := build

# You are going to need to have buffalo installed https://gobuffalo.io/documentation/getting_started/installation/
.PHONY: install
install:
	go get contented
	buffalo plugin install
	yarn install

# Need to fix the docker build, it is pretty old.
.PHONY: build
build:
	docker build .

.PHONY: dev
dev:
	export DIR=`pwd`/mocks/content && buffalo dev

# I would rather use gotestsum, but buffalo does a bunch of DB setup that doesn't play
# nice with go test or gotestsum. Or potentially my tests need some saner / better init
# data around the environment? Should probably move preview into it's own package with how 
# damn slow ffmpeg seek screen tests are on MacOSX.
.PHONY: test
test:
	export DIR=`pwd`/mocks/content && buffalo test ./models ./utils ./managers ./actions

# This works with gotestsum, something about a DB reset is missing or magical Buffalo code.
# The Database side of things doesn't get created with gotestsum yet
# To run one test with gotestsum you can steal this line and pass --run <TestName>
.PHONY: gotestsum
gtest:
	export DIR=`pwd`/mocks/content && gotestsum --format testname ./models
	export DIR=`pwd`/mocks/content && gotestsum --format testname ./utils
	export DIR=`pwd`/mocks/content && buffalo test ./managers
	export DIR=`pwd`/mocks/content && buffalo test ./actions

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

