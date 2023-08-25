# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/
FROM gobuffalo/buffalo:v1.0.1 as builder

ENV GOPROXY http://proxy.golang.org

RUN mkdir -p /src/contented
WORKDIR /src/contented

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

ADD . .
RUN buffalo build --static -o --skip-assets /bin/app

#======================================================================================
# Build out the angular and front end code
#======================================================================================
FROM node:18 as angular

RUN mkdir /contented
WORKDIR /contented
ADD . .

# Clear out any Mac or other OS binaries from an external install
RUN rm -rf public && rm -rf node_modules && mkdir -p public
RUN yarn install
RUN make typescript
RUN ls -la /contented/public && ls -la /contented/public/build/index.html

#======================================================================================
# Build out the main hosted container that doesn't have all the build dependencies
#======================================================================================
FROM alpine
RUN apk add --no-cache bash
RUN apk add --no-cache ca-certificates

WORKDIR /bin/

COPY --from=builder /bin/app .

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

# Uncomment to run the migrations before running the binary:
# CMD /bin/app migrate; /bin/app
CMD exec /bin/app
