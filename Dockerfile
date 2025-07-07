# Usa un'immagine Go per la build
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
COPY . .
RUN go build -o app

# Immagine finale minimale
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
CMD ["./app"]