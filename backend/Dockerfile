FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG VERSION=0.2.0-dev
ARG BUILD_DATE=development
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE}" -o server ./cmd/server

FROM alpine:3.24

RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

COPY --from=builder /app/openapi.yaml .

EXPOSE 8080

ENTRYPOINT ["./server"]