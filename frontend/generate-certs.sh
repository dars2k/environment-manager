#!/bin/bash

# Create certificates directory if it doesn't exist
mkdir -p certs

# Generate a private key
openssl genrsa -out certs/nginx.key 2048

# Generate a certificate signing request
openssl req -new -key certs/nginx.key \
    -out certs/nginx.csr \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Generate the self-signed certificate (valid for 365 days)
openssl x509 -req -days 365 \
    -in certs/nginx.csr \
    -signkey certs/nginx.key \
    -out certs/nginx.crt

# Generate Diffie-Hellman parameters for enhanced security
openssl dhparam -out certs/dhparam.pem 2048

# Set appropriate permissions
chmod 600 certs/nginx.key
chmod 644 certs/nginx.crt
chmod 644 certs/dhparam.pem

echo "Self-signed certificates generated successfully in ./certs/"
echo "Files created:"
echo "  - certs/nginx.key (private key)"
echo "  - certs/nginx.crt (certificate)"
echo "  - certs/dhparam.pem (DH parameters)"
