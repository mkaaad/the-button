FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/the-button ./cmd/

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/the-button .

EXPOSE 8080

ENTRYPOINT ["./the-button"]
