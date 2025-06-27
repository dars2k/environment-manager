#!/bin/bash

# Start SSH service
echo "Starting SSH service..."
/usr/sbin/sshd

# Start health check server
echo "Starting health check server on port 8080..."
python3 /app/health_server.py
