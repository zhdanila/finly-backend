FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY .. .

RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/server/main.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /worker .

COPY .env .

ENV ENV=dev

EXPOSE 8080

CMD ["./worker"]
