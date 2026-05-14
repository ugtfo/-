FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o log-parser ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/log-parser .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/data ./data

EXPOSE 8080

CMD ["./log-parser"]
