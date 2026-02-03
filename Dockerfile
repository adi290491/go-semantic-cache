# Stage 1: Build
FROM golang:1.25.4-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o semantic-cache ./cmd/server

# Stage 2: Minimal Runtime
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/semantic-cache .

EXPOSE 8002

CMD ["./semantic-cache"]