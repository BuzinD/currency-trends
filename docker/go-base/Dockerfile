FROM golang:1.23.3-alpine AS builder

RUN apk --no-cache add bash git make gcc gettext musl-dev curl

WORKDIR /usr/local/src