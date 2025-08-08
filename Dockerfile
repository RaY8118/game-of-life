# Stage 1: Build Go server
FROM golang:1.24 AS builder
WORKDIR /app
COPY backend/ .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o server websocket_server.go

# Stage 2: Final image with Nginx + Go
FROM nginx:alpine
WORKDIR /app

# Copy frontend files to Nginx
COPY frontend/ /usr/share/nginx/html

# Copy Go binary
COPY --from=builder /app/server /app/server

# Start script
COPY start.sh /start.sh
RUN chmod +x /start.sh

EXPOSE 80
EXPOSE 8080
CMD ["/start.sh"]
