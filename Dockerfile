FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gograph .

# --- финальный образ ---
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/gograph .
COPY templates ./templates
COPY ui ./ui

EXPOSE 8181

CMD ["./gograph"]