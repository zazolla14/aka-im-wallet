# syntax=docker/dockerfile:1.4

FROM golang:1.24.5-alpine AS builder
RUN apk add --no-cache git

ENV SERVER_DIR=/im-wallet-cronjob
ENV GOPRIVATE=bitbucket.org/interdigital/*
WORKDIR $SERVER_DIR

COPY go.mod go.sum ./
COPY vendor/ ./vendor
COPY . .

ENV GOFLAGS=-mod=vendor

ARG SERVICE_NAME
ARG CMD_SERVICE_DIR
RUN go build -o _output/${SERVICE_NAME} ./cmd/${CMD_SERVICE_DIR}

FROM alpine:latest
RUN apk add --no-cache bash

ENV SERVER_DIR=/im-wallet-cronjob
WORKDIR $SERVER_DIR

ARG SERVICE_NAME
ENV SERVICE_NAME=${SERVICE_NAME}
COPY --from=builder $SERVER_DIR/_output/$SERVICE_NAME $SERVER_DIR/_output/$SERVICE_NAME

ENTRYPOINT ["/im-wallet-cronjob/_output/publisher"]
