#!/bin/sh
# Start Nginx in background

nginx &

# Start Go Websocket server
/app/server