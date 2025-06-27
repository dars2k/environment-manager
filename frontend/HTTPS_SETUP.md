# HTTPS Setup for Frontend

This frontend now supports HTTPS with self-signed certificates by default.

## Default Configuration

The Docker image automatically generates self-signed certificates during the build process. These certificates:
- Are valid for 365 days
- Use RSA 2048-bit encryption
- Include Diffie-Hellman parameters for enhanced security

## How It Works

1. **Automatic HTTPS**: When you start the container, it will serve content on both:
   - Port 80 (HTTP) - redirects to HTTPS
   - Port 443 (HTTPS) - serves the application

2. **Health Check**: The `/health` endpoint is available on both HTTP and HTTPS for container health monitoring.

3. **Security Features**:
   - TLS 1.2 and 1.3 only
   - Strong cipher suites
   - HSTS (HTTP Strict Transport Security) enabled
   - Security headers (X-Frame-Options, X-Content-Type-Options, etc.)

## Using Custom Certificates

If you want to use your own certificates instead of the auto-generated ones:

### Method 1: Generate Certificates Locally

1. Run the certificate generation script:
   ```bash
   cd frontend
   chmod +x generate-certs.sh
   ./generate-certs.sh
   ```

2. Update your docker-compose.yml to mount the certificates:
   ```yaml
   frontend:
     volumes:
       - ./frontend/certs:/etc/nginx/certs:ro
   ```

### Method 2: Use Existing Certificates

1. Create a `certs` directory in the frontend folder:
   ```bash
   mkdir -p frontend/certs
   ```

2. Copy your certificates:
   ```bash
   cp /path/to/your/nginx.crt frontend/certs/
   cp /path/to/your/nginx.key frontend/certs/
   cp /path/to/your/dhparam.pem frontend/certs/  # Optional but recommended
   ```

3. Update docker-compose.yml as shown above.

## Environment Variables

- `FRONTEND_PORT`: HTTP port (default: 80)
- `FRONTEND_HTTPS_PORT`: HTTPS port (default: 443)

## Browser Warning

When using self-signed certificates, browsers will show a security warning. This is expected behavior. To proceed:
- Chrome: Click "Advanced" → "Proceed to localhost (unsafe)"
- Firefox: Click "Advanced" → "Accept the Risk and Continue"
- Safari: Click "Show Details" → "visit this website"

## Production Recommendations

For production environments:
1. Use certificates from a trusted Certificate Authority (CA)
2. Consider using Let's Encrypt for free SSL certificates
3. Implement certificate renewal automation
4. Use a reverse proxy like Traefik or Nginx Proxy Manager for certificate management

## Troubleshooting

1. **Port Already in Use**: If port 443 is already in use, change the HTTPS port in your .env file:
   ```
   FRONTEND_HTTPS_PORT=8443
   ```

2. **Certificate Issues**: Check the nginx error logs:
   ```bash
   docker logs app-env-manager-frontend
   ```

3. **Mixed Content Warnings**: Ensure all resources are loaded over HTTPS by updating API URLs to use `https://` and `wss://` protocols.
