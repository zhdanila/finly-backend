FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/server/main.go

FROM alpine:3.18

RUN apk add --no-cache wget bash ca-certificates

WORKDIR /root/

COPY --from=builder /worker .
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh

RUN wget -O /bin/goose https://github.com/pressly/goose/releases/download/v3.24.1/goose_linux_amd64 && \
    chmod +x /bin/goose

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]

CMD ["./worker"]
