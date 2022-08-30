# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/

#======================================================================================
# Build out the angular and front end code
#======================================================================================
FROM node:16 as angular

RUN mkdir /contented
WORKDIR /contented
ADD . .

# Clear out any Mac or other OS binaries from an external install
RUN rm -rf public && rm -rf node_modules && mkdir -p public
RUN yarn install
RUN yarn run gulp buildDeploy
RUN ls -la /contented/public && ls -la /contented/public/build/index.html

#======================================================================================
# Build out the go binary
#======================================================================================
FROM gobuffalo/buffalo:v0.18.8 as builder

ENV GO111MODULE on
ENV GOPROXY http://proxy.golang.org

RUN mkdir -p /src/src
WORKDIR /src/src

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

ADD . .
RUN buffalo build --static -o /bin/contented


#======================================================================================
# The actual run environment
#======================================================================================
FROM alpine
RUN apk add --no-cache bash
RUN apk add --no-cache ca-certificates

WORKDIR /bin/

COPY --from=builder /bin/contented .

RUN mkdir -p /public
COPY --from=angular /contented/public/ /public/
RUN ls -la /public/

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production
# ENV GO_ENV=development_docker

# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

# TODO: For some reason out of container with no db works fine, in container tries to connect to the DB
# even if no transactions are made.  Something about the config / connection pool needs a tweak.

# Uncomment to run the migrations before running the binary:
# CMD /bin/app migrate; /bin/app
CMD exec /bin/contented
