# golang:1.25-alpine
FROM golang@sha256:ac09a5f469f307e5da71e766b0bd59c9c49ea460a528cc3e6686513d64a6f1fb AS builder

WORKDIR /app

COPY proto ../proto
COPY upload-service/go.mod upload-service/go.sum ./
RUN go mod edit -replace github.com/athandoan/youtube/proto=../proto
RUN go mod download
COPY upload-service/ .

RUN go build -o upload-service ./cmd/server

# alpine:3.23
FROM alpine@sha256:865b95f46d98cf867a156fe4a135ad3fe50d2056aa3f25ed31662dff6da4eb62
WORKDIR /app
COPY --from=builder /app/upload-service .
EXPOSE 50052
CMD ["./upload-service"]

