#!/usr/bin/env python3
import json
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from datetime import datetime

class HealthCheckHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            response = {
                'status': 'healthy',
                'timestamp': datetime.now().isoformat(),
                'service': 'test-environment',
                'uptime': int(time.time() - server_start_time),
                'version': '1.0.0'
            }
            
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/version':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            response = {
                'current': '1.0.0',
                'available': ['1.0.1', '1.1.0', '2.0.0']
            }
            
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/info':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            response = {
                'hostname': 'test-environment',
                'os': 'Linux',
                'platform': 'Docker',
                'environment': 'test'
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
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            data = json.loads(post_data.decode('utf-8'))
            
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            response = {
                'status': 'success',
                'message': f'Upgrade to version {data.get("version", "unknown")} initiated',
                'timestamp': datetime.now().isoformat()
            }
            
            self.wfile.write(json.dumps(response).encode())
            
        else:
            self.send_response(404)
            self.send_header('Content-type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'Not Found')
    
    def log_message(self, format, *args):
        # Override to add timestamp to logs
        print(f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] {format % args}")

if __name__ == '__main__':
    server_start_time = time.time()
    server_address = ('0.0.0.0', 8080)
    httpd = HTTPServer(server_address, HealthCheckHandler)
    print(f'Starting health check server on {server_address[0]}:{server_address[1]}')
    httpd.serve_forever()
