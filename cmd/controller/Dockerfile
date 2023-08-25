FROM golang:1.21-alpine AS builder
WORKDIR /app

ARG TAG

COPY go.mod go.sum ./
COPY . ./

RUN apk add --no-cache make build-base linux-headers
RUN go mod download
RUN make controller

FROM alpine
WORKDIR /app
COPY --from=builder /app/bin/k8s-image-warden-controller .

ENTRYPOINT [ "./k8s-image-warden-controller" ]