FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM alpine:3.24

RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

COPY --from=builder /app/openapi.yaml .

EXPOSE 8080

ENTRYPOINT ["./server"]