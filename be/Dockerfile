FROM golang:1.24.4 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy && \
    GOOS=linux GOARCH=amd64 go build -o /app/the-button ./cmd/

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/the-button .

EXPOSE 8080

ENTRYPOINT ["./the-button"]