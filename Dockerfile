####################
# Base
FROM golang:1.22-alpine as base
RUN apk add --update --no-cache alpine-sdk

####################
# Deps
FROM base as deps

WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download -x
COPY . /app

####################
# Build
FROM deps as build

WORKDIR /app
RUN go build -tags musl -o /shortener

####################
# shortener
FROM alpine:3 as shortener

ARG ADDR=":8080"

RUN addgroup -S app && adduser -S app -G app
USER app

WORKDIR /app

COPY --from=build /shortener /shortener

HEALTHCHECK --interval=5s --timeout=3s \
    CMD wget --no-verbose --tries=1 --spider http://$ADDR/health || sh -c 'kill -s 15 1 && (sleep 10; kill -s 9 1)'
ENTRYPOINT [ "/shortener" ]