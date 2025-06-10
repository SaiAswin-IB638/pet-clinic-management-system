FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/binary ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/binary .
EXPOSE 8000
CMD ["/app/binary"]