export interface Environment {
  id: string;
  name: string;
  description: string;
  environmentURL: string;
  target: Target;
  credentials: CredentialRef;
  healthCheck: HealthCheckConfig;
  status: Status;
  systemInfo: SystemInfo;
  timestamps: Timestamps;
  commands: CommandConfig;
  upgradeConfig: UpgradeConfig;
  metadata?: Record<string, any>;
}

export interface Target {
  host: string;
  port: number;
  domain?: string;
}

export interface CredentialRef {
  type: 'key' | 'password';
  username: string;
  keyId?: string;
}

export interface HealthCheckConfig {
  enabled: boolean;
  endpoint: string;
  method: string;
  interval: number;
  timeout: number;
  validation: ValidationConfig;
  headers?: Record<string, string>;
}

export interface ValidationConfig {
  type: 'statusCode' | 'jsonRegex';
  value: number | string;
}

export interface Status {
  health: HealthStatus;
  lastCheck: string;
  message: string;
  responseTime: number;
}

export enum HealthStatus {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown',
}

export interface SystemInfo {
  osVersion: string;
  appVersion: string;
  lastUpdated: string;
}

export interface Timestamps {
  createdAt: string;
  updatedAt: string;
  lastRestartAt?: string;
  lastUpgradeAt?: string;
  lastHealthyAt?: string;
}

export interface Credentials {
  type: 'key' | 'password';
  username: string;
}

export interface CreateEnvironmentRequest {
  name: string;
  description: string;
  environmentURL?: string;
  target: Target;
  credentials: Credentials;
  healthCheck: HealthCheckConfig;
  commands: CommandConfig;
  upgradeConfig: UpgradeConfig;
  metadata: Record<string, any>;
}

export interface UpdateEnvironmentRequest {
  name?: string;
  description?: string;
  environmentURL?: string;
  target?: Target;
  credentials?: Credentials;
  healthCheck?: HealthCheckConfig;
  commands?: CommandConfig;
  upgradeConfig?: UpgradeConfig;
  metadata?: Record<string, any>;
}

export interface OperationResponse {
  operationId: string;
  status: string;
  startedAt?: string;
}

export interface RestartRequest {
  force?: boolean;
}

export interface UpgradeRequest {
  version: string;
  backupFirst?: boolean;
  rollbackOnFailure?: boolean;
}

export interface CommandConfig {
  type: CommandType;
  restart: RestartConfig;
}

export type CommandType = 'ssh' | 'http';

export interface RestartConfig {
  enabled: boolean;
  command?: string; // For SSH
  url?: string; // For HTTP
  method?: string; // For HTTP (GET, POST, PUT, PATCH, DELETE)
  headers?: Record<string, string>; // For HTTP
  body?: Record<string, any>; // For HTTP
}

export interface CommandDetails {
  command?: string; // For SSH
  url?: string; // For HTTP
  method?: string; // For HTTP (GET, POST, PUT, PATCH, DELETE)
  headers?: Record<string, string>; // For HTTP
  body?: Record<string, any>; // For HTTP
}

export interface UpgradeConfig {
  enabled: boolean;
  type: CommandType; // SSH or HTTP for upgrade command
  versionListURL: string; // URL to fetch available versions
  versionListMethod?: string; // HTTP method for version list request
  versionListHeaders?: Record<string, string>; // Headers for version list request
  versionListBody?: string; // Body for version list request
  jsonPathResponse: string; // JSONPath to extract version list from response
  upgradeCommand: CommandDetails; // SSH command or HTTP details for upgrade
}

export interface VersionsResponse {
  currentVersion: string;
  availableVersions: string[];
}
