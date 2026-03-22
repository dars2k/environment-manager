# Deployment Guide

## Overview

This guide covers deployment strategies for the Environment Manager in production environments. Docker Compose is the primary recommended approach. Kubernetes examples are provided as a reference for cloud deployments.

## Deployment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   Load Balancer (HTTPS)                      │
│                       (nginx / ALB)                          │
└──────────────────────┬──────────────────────────────────────┘
                       │
          ┌────────────┴────────────┐
          │     Nginx (Frontend)    │
          │  Static files + Proxy   │
          └────────────┬────────────┘
                       │ /api/* , /ws
          ┌────────────┴────────────┐
          │    Go API Server        │
          │    (Backend)            │
          └────────────┬────────────┘
                       │
          ┌────────────┴────────────┐
          │     MongoDB 8.2         │
          │  (Replica Set / Atlas)  │
          └─────────────────────────┘
```

## Docker Compose (Recommended)

### Production Setup

1. Copy and configure `.env`:

```bash
cp .env.example .env
```

2. Set secure values for production:

```bash
# Generate strong secrets
JWT_SECRET=$(openssl rand -hex 32)
SSH_KEY_ENCRYPTION_KEY=$(openssl rand -hex 16)   # exactly 32 bytes

# Set real domain
ALLOWED_ORIGINS=https://your-domain.com
VITE_API_URL=https://your-domain.com/api
VITE_WS_URL=wss://your-domain.com/ws
```

3. Build and start:

```bash
docker compose build
docker compose up -d
```

4. Retrieve the auto-generated admin password:

```bash
docker compose logs backend | grep -i "admin password"
```

### Container Images

**Backend:**
```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/config ./config
EXPOSE 8080
CMD ["./server"]
```

**Frontend:**
```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:stable-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

---

## Kubernetes (Reference)

> These examples are a starting point. Adapt resource limits, replica counts, and secrets to your cluster.

### Backend Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: env-manager-backend
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
        image: ghcr.io/dars2k/environment-manager-backend:latest
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
        - name: SSH_KEY_ENCRYPTION_KEY
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: ssh-encryption-key
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health
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

### Frontend Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: env-manager-frontend
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
        image: ghcr.io/dars2k/environment-manager-frontend:latest
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

### Ingress with WebSocket Support

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: env-manager-ingress
  namespace: production
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
spec:
  tls:
  - hosts:
    - your-domain.com
    secretName: env-manager-tls
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: backend-service
            port:
              number: 80
      - path: /ws
        pathType: Prefix
        backend:
          service:
            name: backend-service
            port:
              number: 80
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 80
```

### Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: production
type: Opaque
stringData:
  mongodb-uri: "mongodb://user:pass@mongodb:27017/app-env-manager?authSource=admin"
  jwt-secret: "<openssl rand -hex 32>"
  ssh-encryption-key: "<openssl rand -hex 16>"
```

### Horizontal Pod Autoscaling

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
    name: env-manager-backend
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

---

## Database

### MongoDB Atlas (Recommended for Production)

```bash
# Create cluster
atlas clusters create env-manager \
  --provider AWS --region US_EAST_1 \
  --tier M10 --members 3

# Create user
atlas dbusers create \
  --username env-manager \
  --password <secure-password> \
  --role readWrite@app-env-manager
```

Set `MONGODB_URI` in your `.env` or Kubernetes secret to the Atlas connection string.

### Self-Hosted Replica Set

Use a MongoDB replica set for production to enable high availability. The minimum recommended setup is a 3-member replica set.

```yaml
apiVersion: apps/v1
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
        image: mongo:8.2
        command: ["mongod", "--replSet", "rs0", "--bind_ip_all"]
        ports:
        - containerPort: 27017
        volumeMounts:
        - name: mongo-data
          mountPath: /data/db
  volumeClaimTemplates:
  - metadata:
      name: mongo-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
```

**Recommended indexes:**

```javascript
db.environments.createIndex({ "name": 1 }, { unique: true })
db.environments.createIndex({ "status.health": 1 })
db.logs.createIndex({ "timestamp": -1, "environmentId": 1 })
db.logs.createIndex({ "userId": 1, "timestamp": -1 })
```

---

## Backup and Restore

### Automated Backup Script

```bash
#!/bin/bash
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups/mongodb-${TIMESTAMP}"

mongodump \
  --uri="${MONGODB_URI}" \
  --out="${BACKUP_DIR}" \
  --gzip

# Upload to S3
aws s3 sync "${BACKUP_DIR}" "s3://your-backup-bucket/mongodb-${TIMESTAMP}/"
```

### Recovery Objectives

| Objective | Target |
|-----------|--------|
| RTO (Recovery Time) | < 4 hours |
| RPO (Recovery Point) | < 1 hour |

---

## Security Hardening

1. **Use HTTPS** — terminate TLS at the load balancer or nginx
2. **Strong secrets** — generate `JWT_SECRET` and `SSH_KEY_ENCRYPTION_KEY` with `openssl rand`
3. **Restrict ALLOWED_ORIGINS** — list only your actual frontend domains
4. **Non-root containers** — all containers run as non-root users
5. **Network isolation** — use Kubernetes NetworkPolicy or Docker bridge networks to restrict inter-service traffic
6. **Secret rotation** — rotate `JWT_SECRET` periodically (invalidates existing sessions)

---

## CI/CD

The repository includes GitHub Actions workflows for CI. For CD, adapt this pattern:

```yaml
name: Deploy

on:
  push:
    branches: [master]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Run tests
      run: |
        cd backend && go test ./...
        cd frontend && npm ci && npm test -- --run

    - name: Build and push images
      run: |
        docker build -t ghcr.io/dars2k/environment-manager-backend:${{ github.sha }} ./backend
        docker build -t ghcr.io/dars2k/environment-manager-frontend:${{ github.sha }} ./frontend
        docker push ghcr.io/dars2k/environment-manager-backend:${{ github.sha }}
        docker push ghcr.io/dars2k/environment-manager-frontend:${{ github.sha }}

    - name: Deploy
      run: |
        # Update deployment with new image tags
        kubectl set image deployment/env-manager-backend \
          backend=ghcr.io/dars2k/environment-manager-backend:${{ github.sha }} -n production
        kubectl set image deployment/env-manager-frontend \
          frontend=ghcr.io/dars2k/environment-manager-frontend:${{ github.sha }} -n production
```
