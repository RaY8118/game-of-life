# Stage 1: Build Go server
FROM golang:1.24 AS builder
WORKDIR /app
COPY backend/ .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o server websocket_server.go

# Stage 2: Final image (no nginx!)
FROM alpine:latest
WORKDIR /app

# Copy frontend and server
COPY frontend/ ./frontend
COPY --from=builder /app/server /app/server

EXPOSE 8080
ENV PORT=8080

CMD ["/app/server"]
