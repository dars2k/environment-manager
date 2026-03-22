import { describe, it, expect, vi, beforeEach } from 'vitest';
import axios from 'axios';
import { environmentApi } from '../environments';
import { HealthStatus } from '@/types/environment';

vi.mock('axios');
const mockedAxios = vi.mocked(axios);

const mockEnvironment = {
  id: 'env-1',
  name: 'Test Env',
  description: 'desc',
  target: { host: 'localhost', port: 22 },
  credentials: { type: 'password', username: 'user' },
  healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
  status: { health: HealthStatus.Unknown, lastCheck: new Date().toISOString(), message: '', responseTime: 0 },
  systemInfo: { osVersion: '', appVersion: '', lastUpdated: new Date().toISOString() },
  timestamps: { createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  commands: { type: 'ssh', restart: { enabled: false } },
  upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
};

describe('environmentApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('list returns environments array', async () => {
    const mockData = { environments: [mockEnvironment], pagination: { page: 1, limit: 10, total: 1 } };
    mockedAxios.get = vi.fn().mockResolvedValue({ data: { data: mockData } });

    const result = await environmentApi.list();
    expect(result.environments).toHaveLength(1);
    expect(result.environments[0].id).toBe('env-1');
  });

  it('list handles null environments by returning empty array', async () => {
    const mockData = { environments: null, pagination: { page: 1, limit: 10, total: 0 } };
    mockedAxios.get = vi.fn().mockResolvedValue({ data: { data: mockData } });

    const result = await environmentApi.list();
    expect(result.environments).toEqual([]);
  });

  it('get returns environment by id', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({ data: { data: { environment: mockEnvironment } } });

    const result = await environmentApi.get('env-1');
    expect(result.id).toBe('env-1');
  });

  it('create posts and returns environment', async () => {
    mockedAxios.post = vi.fn().mockResolvedValue({ data: { data: mockEnvironment } });

    const result = await environmentApi.create({ name: 'New Env', target: { host: 'localhost', port: 22 }, credentials: { type: 'password', username: 'user' }, healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } }, commands: { type: 'ssh', restart: { enabled: false } }, upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} } });
    expect(result.name).toBe('Test Env');
  });

  it('update puts and returns environment', async () => {
    mockedAxios.put = vi.fn().mockResolvedValue({ data: { data: { ...mockEnvironment, name: 'Updated' } } });

    const result = await environmentApi.update('env-1', { name: 'Updated' });
    expect(result.name).toBe('Updated');
  });

  it('delete calls delete endpoint', async () => {
    mockedAxios.delete = vi.fn().mockResolvedValue({});

    await expect(environmentApi.delete('env-1')).resolves.toBeUndefined();
    expect(mockedAxios.delete).toHaveBeenCalledWith('/api/v1/environments/env-1');
  });

  it('restart posts force flag', async () => {
    const opResponse = { operationId: 'op-1', status: 'success', message: 'done' };
    mockedAxios.post = vi.fn().mockResolvedValue({ data: { data: opResponse } });

    const result = await environmentApi.restart('env-1', true);
    expect(result.operationId).toBe('op-1');
    expect(mockedAxios.post).toHaveBeenCalledWith('/api/v1/environments/env-1/restart', { force: true });
  });

  it('restart defaults force to false', async () => {
    mockedAxios.post = vi.fn().mockResolvedValue({ data: { data: { operationId: 'op-2' } } });

    await environmentApi.restart('env-1');
    expect(mockedAxios.post).toHaveBeenCalledWith('/api/v1/environments/env-1/restart', { force: false });
  });

  it('checkHealth posts to check-health endpoint', async () => {
    mockedAxios.post = vi.fn().mockResolvedValue({});

    await expect(environmentApi.checkHealth('env-1')).resolves.toBeUndefined();
    expect(mockedAxios.post).toHaveBeenCalledWith('/api/v1/environments/env-1/check-health');
  });

  it('getVersions returns versions response', async () => {
    const versionsData = { availableVersions: ['v1', 'v2'], currentVersion: 'v1' };
    mockedAxios.get = vi.fn().mockResolvedValue({ data: { data: versionsData } });

    const result = await environmentApi.getVersions('env-1');
    expect(result.availableVersions).toContain('v1');
  });

  it('upgrade posts version and returns operation response', async () => {
    const opResponse = { operationId: 'op-3', status: 'started', message: 'upgrading' };
    mockedAxios.post = vi.fn().mockResolvedValue({ data: { data: opResponse } });

    const result = await environmentApi.upgrade('env-1', { version: 'v2' });
    expect(result.operationId).toBe('op-3');
  });
});
