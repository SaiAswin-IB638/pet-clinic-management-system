FROM golang:1.24.3-alpine AS builder
RUN apk add --no-cache tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/binary ./cmd/api/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/binary .
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
EXPOSE 8000
CMD ["/app/binary"]