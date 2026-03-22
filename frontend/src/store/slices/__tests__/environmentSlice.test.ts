import { describe, it, expect } from 'vitest';
import { configureStore } from '@reduxjs/toolkit';
import reducer, {
  setEnvironments,
  addEnvironment,
  updateEnvironment,
  removeEnvironment,
  setSelectedEnvironment,
  updateEnvironmentStatus,
  setLoading,
  setError,
} from '../environmentSlice';
import { Environment, HealthStatus } from '@/types/environment';

const makeEnv = (id: string): Environment => ({
  id,
  name: `env-${id}`,
  description: 'test',
  target: { host: 'localhost', port: 22 },
  credentials: { type: 'password', username: 'user' },
  healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
  status: { health: HealthStatus.Unknown, lastCheck: new Date().toISOString(), message: '', responseTime: 0 },
  systemInfo: { osVersion: '', appVersion: '', lastUpdated: new Date().toISOString() },
  timestamps: { createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  commands: { type: 'ssh', restart: { enabled: false } },
  upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
});

describe('environmentSlice', () => {
  const initialState = {
    environments: [],
    selectedEnvironment: null,
    loading: false,
    error: null,
  };

  it('returns initial state', () => {
    expect(reducer(undefined, { type: '@@INIT' })).toEqual(initialState);
  });

  it('setEnvironments replaces list and clears error', () => {
    const envs = [makeEnv('1'), makeEnv('2')];
    const state = reducer({ ...initialState, error: 'old error' }, setEnvironments(envs));
    expect(state.environments).toHaveLength(2);
    expect(state.error).toBeNull();
  });

  it('addEnvironment appends to list', () => {
    const env = makeEnv('1');
    const state = reducer(initialState, addEnvironment(env));
    expect(state.environments).toHaveLength(1);
    expect(state.environments[0].id).toBe('1');
  });

  it('updateEnvironment replaces existing entry by id', () => {
    const env = makeEnv('1');
    const stateWithEnv = { ...initialState, environments: [env] };
    const updated = { ...env, name: 'updated-name' };
    const state = reducer(stateWithEnv, updateEnvironment(updated));
    expect(state.environments[0].name).toBe('updated-name');
  });

  it('updateEnvironment also updates selectedEnvironment when IDs match', () => {
    const env = makeEnv('1');
    const stateWithSelected = { ...initialState, environments: [env], selectedEnvironment: env };
    const updated = { ...env, name: 'new-name' };
    const state = reducer(stateWithSelected, updateEnvironment(updated));
    expect(state.selectedEnvironment?.name).toBe('new-name');
  });

  it('updateEnvironment does nothing when id not found', () => {
    const env = makeEnv('1');
    const stateWithEnv = { ...initialState, environments: [env] };
    const other = makeEnv('99');
    const state = reducer(stateWithEnv, updateEnvironment(other));
    expect(state.environments).toHaveLength(1);
    expect(state.environments[0].id).toBe('1');
  });

  it('removeEnvironment filters out by id', () => {
    const envs = [makeEnv('1'), makeEnv('2')];
    const state = reducer({ ...initialState, environments: envs }, removeEnvironment('1'));
    expect(state.environments).toHaveLength(1);
    expect(state.environments[0].id).toBe('2');
  });

  it('removeEnvironment clears selectedEnvironment when matching', () => {
    const env = makeEnv('1');
    const stateWithSelected = { ...initialState, environments: [env], selectedEnvironment: env };
    const state = reducer(stateWithSelected, removeEnvironment('1'));
    expect(state.selectedEnvironment).toBeNull();
  });

  it('removeEnvironment does not clear selectedEnvironment when non-matching', () => {
    const env1 = makeEnv('1');
    const env2 = makeEnv('2');
    const stateWithSelected = { ...initialState, environments: [env1, env2], selectedEnvironment: env2 };
    const state = reducer(stateWithSelected, removeEnvironment('1'));
    expect(state.selectedEnvironment?.id).toBe('2');
  });

  it('setSelectedEnvironment sets and clears', () => {
    const env = makeEnv('1');
    let state = reducer(initialState, setSelectedEnvironment(env));
    expect(state.selectedEnvironment?.id).toBe('1');
    state = reducer(state, setSelectedEnvironment(null));
    expect(state.selectedEnvironment).toBeNull();
  });

  it('updateEnvironmentStatus updates status on matching env', () => {
    const env = makeEnv('1');
    const stateWithEnv = { ...initialState, environments: [env] };
    const newStatus = { health: HealthStatus.Healthy, lastCheck: new Date().toISOString(), message: 'ok', responseTime: 100 };
    const state = reducer(stateWithEnv, updateEnvironmentStatus({ id: '1', status: newStatus }));
    expect(state.environments[0].status.health).toBe(HealthStatus.Healthy);
  });

  it('updateEnvironmentStatus also updates selectedEnvironment', () => {
    const env = makeEnv('1');
    const stateWithSelected = { ...initialState, environments: [env], selectedEnvironment: env };
    const newStatus = { health: HealthStatus.Unhealthy, lastCheck: new Date().toISOString(), message: 'err', responseTime: 0 };
    const state = reducer(stateWithSelected, updateEnvironmentStatus({ id: '1', status: newStatus }));
    expect(state.selectedEnvironment?.status.health).toBe(HealthStatus.Unhealthy);
  });

  it('updateEnvironmentStatus does nothing when id not found', () => {
    const env = makeEnv('1');
    const stateWithEnv = { ...initialState, environments: [env] };
    const state = reducer(stateWithEnv, updateEnvironmentStatus({ id: '99', status: {} }));
    expect(state.environments[0].status.health).toBe(HealthStatus.Unknown);
  });

  it('setLoading sets loading state', () => {
    let state = reducer(initialState, setLoading(true));
    expect(state.loading).toBe(true);
    state = reducer(state, setLoading(false));
    expect(state.loading).toBe(false);
  });

  it('setError sets error state', () => {
    let state = reducer(initialState, setError('something broke'));
    expect(state.error).toBe('something broke');
    state = reducer(state, setError(null));
    expect(state.error).toBeNull();
  });

  it('works with configureStore', () => {
    const store = configureStore({ reducer: { environments: reducer } });
    store.dispatch(addEnvironment(makeEnv('x')));
    expect(store.getState().environments.environments).toHaveLength(1);
  });
});
