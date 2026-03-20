#!/usr/bin/env python3
import json
import os
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from datetime import datetime

VERSION_FILE = '/tmp/app_version'

def get_current_version():
    if os.path.exists(VERSION_FILE):
        with open(VERSION_FILE, 'r') as f:
            return f.read().strip()
    return os.environ.get('APP_VERSION', '1.0.0')

def set_current_version(version):
    with open(VERSION_FILE, 'w') as f:
        f.write(version.strip())

# Initialize version file from env var
if not os.path.exists(VERSION_FILE):
    set_current_version(os.environ.get('APP_VERSION', '1.0.0'))

server_start_time = time.time()
health_state = {'healthy': True, 'response_delay_ms': 0}

AVAILABLE_VERSIONS = ['1.0.1', '1.1.0', '2.0.0', '2.1.0']


class HealthCheckHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if health_state['response_delay_ms'] > 0:
            time.sleep(health_state['response_delay_ms'] / 1000.0)

        if self.path == '/health':
            if health_state['healthy']:
                self.send_response(200)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                response = {
                    'status': 'healthy',
                    'timestamp': datetime.now().isoformat(),
                    'service': 'test-environment',
                    'uptime': int(time.time() - server_start_time),
                    'version': get_current_version()
                }
            else:
                self.send_response(503)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                response = {
                    'status': 'unhealthy',
                    'timestamp': datetime.now().isoformat(),
                    'service': 'test-environment',
                }
            self.wfile.write(json.dumps(response).encode())

        elif self.path == '/version':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            current = get_current_version()
            # Exclude current version from available list
            available = [v for v in AVAILABLE_VERSIONS if v != current]
            response = {
                'current': current,
                'available': available
            }
            self.wfile.write(json.dumps(response).encode())

        elif self.path == '/info':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = {
                'hostname': os.environ.get('HOSTNAME', 'test-environment'),
                'os': 'Linux',
                'platform': 'Docker',
                'environment': 'test',
                'version': get_current_version()
            }
            self.wfile.write(json.dumps(response).encode())

        elif self.path == '/metrics':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = {
                'cpu_usage': 15.2,
                'memory_usage': 45.7,
                'disk_usage': 23.1,
                'active_connections': 3
            }
            self.wfile.write(json.dumps(response).encode())

        else:
            self.send_response(404)
            self.send_header('Content-type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'Not Found')

    def do_POST(self):
        if self.path == '/restart':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = {
                'status': 'success',
                'message': 'Service restart initiated',
                'timestamp': datetime.now().isoformat()
            }
            self.wfile.write(json.dumps(response).encode())

        elif self.path == '/upgrade':
            content_length = int(self.headers.get('Content-Length', 0))
            post_data = self.rfile.read(content_length) if content_length > 0 else b'{}'
            data = json.loads(post_data.decode('utf-8'))
            new_version = data.get('version', get_current_version())
            set_current_version(new_version)
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = {
                'status': 'success',
                'message': f'Upgrade to version {new_version} completed',
                'version': new_version,
                'timestamp': datetime.now().isoformat()
            }
            self.wfile.write(json.dumps(response).encode())

        elif self.path == '/toggle-health':
            health_state['healthy'] = not health_state['healthy']
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = {
                'status': 'success',
                'healthy': health_state['healthy'],
                'message': f"Health toggled to {'healthy' if health_state['healthy'] else 'unhealthy'}"
            }
            self.wfile.write(json.dumps(response).encode())

        elif self.path == '/set-response-time':
            content_length = int(self.headers.get('Content-Length', 0))
            post_data = self.rfile.read(content_length) if content_length > 0 else b'{}'
            data = json.loads(post_data.decode('utf-8'))
            health_state['response_delay_ms'] = int(data.get('ms', 0))
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = {
                'status': 'success',
                'response_delay_ms': health_state['response_delay_ms']
            }
            self.wfile.write(json.dumps(response).encode())

        else:
            self.send_response(404)
            self.send_header('Content-type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'Not Found')

    def log_message(self, format, *args):
        print(f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] {format % args}")


if __name__ == '__main__':
    port = int(os.environ.get('HTTP_PORT', 8080))
    server_address = ('0.0.0.0', port)
    httpd = HTTPServer(server_address, HealthCheckHandler)
    print(f'Starting health check server on port {port}')
    print(f'Initial version: {get_current_version()}')
    httpd.serve_forever()
