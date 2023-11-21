# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION="1.0.0"
ARG BUILDNUM=""

# Build Geth in a stock Go builder container
FROM golang:1.21-alpine as builder

RUN apk update --no-cache && apk add --no-cache tzdata linux-headers

# Get dependencies - will also be cached if we won't change go.mod/go.sum
COPY go.mod /mosaic/
COPY go.sum /mosaic/
RUN cd /mosaic && go mod download

ADD . /mosaic

RUN cd /mosaic && go build -ldflags "-s -w" -o build/mosaic

# Pull Geth into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /mosaic/build/mosaic /usr/local/bin/

EXPOSE 25010
ENTRYPOINT ["mosaic"]

# Add some metadata labels to help programatic image consumption
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

LABEL commit="$COMMIT" version="$VERSION" buildnum="$BUILDNUM"
