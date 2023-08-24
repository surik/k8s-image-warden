FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
COPY . ./

RUN apk add --no-cache make
RUN go mod download
RUN make agent

FROM alpine
WORKDIR /app
COPY --from=builder /app/bin/k8s-image-warden-agent .

ENTRYPOINT [ "./k8s-image-warden-agent" ]