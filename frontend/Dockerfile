# Build stage
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package.json ./

# Install dependencies with clean state and proper esbuild
RUN apk add --no-cache python3 make g++ && \
    rm -rf node_modules package-lock.json ~/.npm && \
    npm cache clean --force && \
    npm install && \
    npm uninstall esbuild && \
    npm install esbuild@0.21.5

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Certificate generation stage (optional)
FROM alpine:latest AS cert-generator

# Install OpenSSL
RUN apk add --no-cache openssl

# Create certificates directory
WORKDIR /certs

# Generate self-signed certificates
RUN openssl genrsa -out nginx.key 2048 && \
    openssl req -new -key nginx.key \
        -out nginx.csr \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost" && \
    openssl x509 -req -days 365 \
        -in nginx.csr \
        -signkey nginx.key \
        -out nginx.crt && \
    openssl dhparam -out dhparam.pem 2048 && \
    chmod 600 nginx.key && \
    chmod 644 nginx.crt && \
    chmod 644 dhparam.pem

# Production stage
FROM nginx:alpine

# Install wget for health checks
RUN apk add --no-cache wget

# Copy custom nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built files from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Create certificates directory
RUN mkdir -p /etc/nginx/certs

# Copy self-signed certificates from cert-generator stage
# These will be used if no certificates are mounted
COPY --from=cert-generator /certs/* /etc/nginx/certs/

# Create non-root user
RUN adduser -D -g '' appuser && \
    chown -R appuser:appuser /usr/share/nginx/html && \
    chown -R appuser:appuser /var/cache/nginx && \
    chown -R appuser:appuser /var/log/nginx && \
    chown -R appuser:appuser /etc/nginx/certs && \
    touch /var/run/nginx.pid && \
    chown -R appuser:appuser /var/run/nginx.pid

# Expose ports
EXPOSE 80 443

# Health check using HTTP (since health endpoint is available on both)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
