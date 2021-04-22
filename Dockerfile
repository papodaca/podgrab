ARG GO_VERSION=1.15.2

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk update && apk add alpine-sdk git && rm -rf /var/cache/apk/*

RUN mkdir -p /api
WORKDIR /api

COPY go.mod .
COPY go.sum .
RUN go mod download && \
  go get github.com/jessevdk/go-assets-builder

COPY . .
RUN go generate client/client.go && \
  go build -o ./app ./main.go

FROM alpine:latest

LABEL org.opencontainers.image.source="https://github.com/akhilrex/podgrab"

ENV CONFIG=/config
ENV DATA=/assets
ENV UID=998
ENV PID=100
ENV GIN_MODE=release
VOLUME ["/config", "/assets"]
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN mkdir -p /config; \
    mkdir -p /assets; \
    mkdir -p /api

RUN chmod 777 /config; \
    chmod 777 /assets

WORKDIR /api
COPY --from=builder /api/app .
COPY webassets ./webassets

EXPOSE 8080

ENTRYPOINT ["./app"]