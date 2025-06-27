// MongoDB initialization script
// This script runs when the MongoDB container is first created

// Switch to the app database
db = db.getSiblingDB('app-env-manager');

// Create a dedicated user for the application
db.createUser({
  user: 'appuser',
  pwd: 'apppassword',
  roles: [
    {
      role: 'readWrite',
      db: 'app-env-manager'
    }
  ]
});

// Create collections with schema validation
db.createCollection('environments', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['name', 'target', 'credentials', 'healthCheck', 'status', 'timestamps'],
      properties: {
        name: {
          bsonType: 'string',
          description: 'Environment name must be a string'
        },
        target: {
          bsonType: 'object',
          required: ['host', 'port'],
          properties: {
            host: { bsonType: 'string' },
            port: { bsonType: 'int' },
            domain: { bsonType: 'string' }
          }
        },
        status: {
          bsonType: 'object',
          required: ['health', 'lastCheck'],
          properties: {
            health: {
              enum: ['healthy', 'unhealthy', 'unknown']
            }
          }
        }
      }
    }
  }
});

db.createCollection('audit_log', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['timestamp', 'environmentId', 'type', 'severity', 'actor', 'action'],
      properties: {
        timestamp: {
          bsonType: 'date',
          description: 'Timestamp must be a date'
        },
        type: {
          enum: ['health_change', 'restart', 'shutdown', 'upgrade', 'config_update', 'operation_start', 'operation_complete', 'operation_failed']
        },
        severity: {
          enum: ['debug', 'info', 'warning', 'error', 'critical']
        }
      }
    }
  }
});

// Create indexes for better performance
db.environments.createIndex({ 'name': 1 }, { unique: true });
db.environments.createIndex({ 'status.health': 1 });
db.environments.createIndex({ 'timestamps.lastCheck': 1 });

db.audit_log.createIndex({ 'timestamp': -1, 'environmentId': 1 });
db.audit_log.createIndex({ 'type': 1 });
db.audit_log.createIndex({ 'actor.id': 1 });
db.audit_log.createIndex({ 'tags': 1 });

print('MongoDB initialization completed successfully');
