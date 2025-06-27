# Deployment Guide

## Overview

This guide covers deployment strategies, infrastructure requirements, and best practices for deploying the Application Environment Manager in production environments.

## Deployment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Load Balancer (HTTPS)                     │
│                         (nginx/ALB)                          │
└──────────────────────┬────────────────┬────────────────────┘
                       │                │
        ┌──────────────┴───┐    ┌──────┴──────────────┐
        │   Frontend CDN   │    │   API Gateway       │
        │   (CloudFront)   │    │   (nginx/Kong)      │
        └──────────────────┘    └─────────┬───────────┘
                                         │
                          ┌──────────────┴──────────────┐
                          │      Backend Cluster        │
                          │  ┌──────┐ ┌──────┐ ┌──────┐│
                          │  │ App  │ │ App  │ │ App  ││
                          │  │ Pod  │ │ Pod  │ │ Pod  ││
                          │  └──────┘ └──────┘ └──────┘│
                          └──────────────┬──────────────┘
                                        │
                          ┌─────────────┴──────────────┐
                          │      MongoDB Cluster       │
                          │   (Replica Set/Atlas)      │
                          └────────────────────────────┘
```

## Container Strategy

### 1. Frontend Dockerfile

```dockerfile
# Build stage
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### 2. Backend Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
EXPOSE 8080
CMD ["./main"]
```

### 3. Docker Compose (Development)

```yaml
version: '3.8'

services:
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.dev
    ports:
      - "3000:3000"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    environment:
      - REACT_APP_API_URL=http://localhost:8080/api/v1
      - REACT_APP_WS_URL=ws://localhost:8080/api/v1

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    volumes:
      - ./backend:/app
    environment:
      - MONGODB_URI=mongodb://mongo:27017/app-env-manager
      - JWT_SECRET=development-secret
      - SSH_KEY_ENCRYPTION_KEY=development-key
    depends_on:
      - mongo

  mongo:
    image: mongo:6
    ports:
      - "27017:27017"
    volumes:
      - mongo-data:/data/db
    environment:
      - MONGO_INITDB_DATABASE=app-env-manager

volumes:
  mongo-data:
```

## Kubernetes Deployment

### 1. Backend Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-env-manager-backend
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: your-registry/app-env-manager-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: MONGODB_URI
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: mongodb-uri
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: jwt-secret
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: backend-service
  namespace: production
spec:
  selector:
    app: backend
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

### 2. Frontend Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-env-manager-frontend
  namespace: production
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: frontend
        image: your-registry/app-env-manager-frontend:latest
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
  namespace: production
spec:
  selector:
    app: frontend
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP
```

### 3. Ingress Configuration

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-env-manager-ingress
  namespace: production
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/websocket-services: backend-service
spec:
  tls:
  - hosts:
    - app-env-manager.example.com
    - api.app-env-manager.example.com
    secretName: app-env-manager-tls
  rules:
  - host: app-env-manager.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 80
  - host: api.app-env-manager.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: backend-service
            port:
              number: 80
```

## Database Deployment

### MongoDB Atlas (Recommended for Production)

1. **Create Cluster**
   ```bash
   # Using MongoDB Atlas CLI
   atlas clusters create app-env-manager \
     --provider AWS \
     --region US_EAST_1 \
     --tier M10 \
     --members 3 \
     --diskSizeGB 10
   ```

2. **Configure Network Access**
   ```bash
   # Add IP whitelist
   atlas accessLists create \
     --type ipAddress \
     --comment "Kubernetes cluster" \
     --ipAddress "x.x.x.x/32"
   ```

3. **Create Database User**
   ```bash
   atlas dbusers create \
     --username app-env-manager \
     --password <secure-password> \
     --role readWrite@app-env-manager
   ```

### Self-Hosted MongoDB

```yaml
apiVersion: v1
kind: StatefulSet
metadata:
  name: mongodb
  namespace: production
spec:
  serviceName: mongodb-service
  replicas: 3
  selector:
    matchLabels:
      app: mongodb
  template:
    metadata:
      labels:
        app: mongodb
    spec:
      containers:
      - name: mongodb
        image: mongo:6
        command:
          - mongod
          - "--replSet"
          - rs0
          - "--bind_ip_all"
        ports:
        - containerPort: 27017
        volumeMounts:
        - name: mongo-persistent-storage
          mountPath: /data/db
  volumeClaimTemplates:
  - metadata:
      name: mongo-persistent-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

## Environment Configuration

### 1. ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: production
data:
  config.yaml: |
    server:
      host: 0.0.0.0
      port: 8080
      readTimeout: 30s
      writeTimeout: 30s
    
    database:
      database: app-env-manager
      maxConnections: 100
      timeout: 10s
    
    health:
      checkInterval: 30s
      timeout: 5s
      maxRetries: 3
    
    ssh:
      connectionTimeout: 30s
      commandTimeout: 300s
      maxConnections: 50
```

### 2. Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: production
type: Opaque
stringData:
  mongodb-uri: "mongodb://username:password@mongodb:27017/app-env-manager"
  jwt-secret: "your-secure-jwt-secret"
  ssh-encryption-key: "your-32-byte-encryption-key"
```

## CI/CD Pipeline

### GitHub Actions Example

```yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Run Backend Tests
      run: |
        cd backend
        go test ./...
    
    - name: Run Frontend Tests
      run: |
        cd frontend
        npm ci
        npm test

  build-and-deploy:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Build Backend Docker Image
      run: |
        docker build -t ${{ secrets.REGISTRY }}/app-env-manager-backend:${{ github.sha }} ./backend
        docker push ${{ secrets.REGISTRY }}/app-env-manager-backend:${{ github.sha }}
    
    - name: Build Frontend Docker Image
      run: |
        docker build -t ${{ secrets.REGISTRY }}/app-env-manager-frontend:${{ github.sha }} ./frontend
        docker push ${{ secrets.REGISTRY }}/app-env-manager-frontend:${{ github.sha }}
    
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/app-env-manager-backend \
          backend=${{ secrets.REGISTRY }}/app-env-manager-backend:${{ github.sha }} \
          -n production
        
        kubectl set image deployment/app-env-manager-frontend \
          frontend=${{ secrets.REGISTRY }}/app-env-manager-frontend:${{ github.sha }} \
          -n production
```

## Scaling Strategies

### 1. Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: app-env-manager-backend
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### 2. Database Scaling

- **Read Replicas**: Configure MongoDB replica sets for read scaling
- **Sharding**: Implement sharding for large datasets
- **Connection Pooling**: Optimize connection pool sizes

### 3. Caching Strategy

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-cache
  namespace: production
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
```

## Monitoring Setup

### 1. Prometheus Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    
    scrape_configs:
    - job_name: 'app-env-manager'
      kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - production
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: backend
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: instance
```

### 2. Grafana Dashboards

Create dashboards for:
- API response times and error rates
- WebSocket connection metrics
- SSH operation success/failure rates
- Environment health status overview
- Resource utilization (CPU, memory, disk)

## Backup and Disaster Recovery

### 1. Database Backups

```bash
#!/bin/bash
# Automated backup script

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"

# MongoDB backup
mongodump \
  --uri="$MONGODB_URI" \
  --out="$BACKUP_DIR/mongodb-$TIMESTAMP" \
  --gzip

# Upload to S3
aws s3 cp \
  "$BACKUP_DIR/mongodb-$TIMESTAMP" \
  "s3://your-backup-bucket/mongodb-$TIMESTAMP" \
  --recursive
```

### 2. Disaster Recovery Plan

1. **RTO (Recovery Time Objective)**: < 4 hours
2. **RPO (Recovery Point Objective)**: < 1 hour
3. **Multi-region deployment** for high availability
4. **Automated failover** using health checks
5. **Regular disaster recovery drills**

## Security Hardening

1. **Network Policies**
   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: backend-network-policy
     namespace: production
   spec:
     podSelector:
       matchLabels:
         app: backend
     policyTypes:
     - Ingress
     - Egress
     ingress:
     - from:
       - podSelector:
           matchLabels:
             app: frontend
       ports:
       - protocol: TCP
         port: 8080
   ```

2. **Pod Security Policies**
   - Run as non-root user
   - Read-only root filesystem
   - No privileged containers

3. **Secrets Management**
   - Use Kubernetes Secrets
   - Consider HashiCorp Vault for advanced use cases
   - Enable encryption at rest

## Performance Optimization

1. **CDN Configuration**
   - Cache static assets
   - Compress responses
   - Geographic distribution

2. **Database Indexes**
   ```javascript
   // Create indexes for optimal performance
   db.environments.createIndex({ "name": 1 }, { unique: true })
   db.environments.createIndex({ "status.health": 1 })
   db.audit_log.createIndex({ "timestamp": -1, "environmentId": 1 })
   ```

3. **Resource Limits**
   - Set appropriate CPU and memory limits
   - Monitor and adjust based on usage patterns
   - Use quality of service (QoS) classes
