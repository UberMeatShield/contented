.DEFAULT_GOAL := build

.PHONY: build
build:
	docker build .

# I would rather use gotestsum, but buffalo does a bunch of DB setup that doesn't play
# nice with go test or gotestsum. Or potentially my tests need some saner / better init
# data around the environment? Should probably move preview into it's own package with how 
# damn slow ffmpeg tests are on MacOSX.
.PHONY: test
test:
	export DIR=`pwd`/mocks/content && buffalo test
	#export DIR=`pwd`/mocks/content && GO_ENV="test" && gotestsum --format testname ./models ./managers ./actions ./utils
