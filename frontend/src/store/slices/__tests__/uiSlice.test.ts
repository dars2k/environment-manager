import { describe, it, expect, vi } from 'vitest';
import reducer, {
  toggleSidebar,
  setSidebarOpen,
  toggleDarkMode,
  setEnvironmentCreateDialogOpen,
  setEnvironmentEditDialogOpen,
  setSelectedEnvironment,
  openConfirmDialog,
  closeConfirmDialog,
} from '../uiSlice';
import { Environment, HealthStatus } from '@/types/environment';

const makeEnv = (): Environment => ({
  id: 'env-1',
  name: 'env',
  description: '',
  target: { host: 'localhost', port: 22 },
  credentials: { type: 'password', username: 'user' },
  healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
  status: { health: HealthStatus.Unknown, lastCheck: new Date().toISOString(), message: '', responseTime: 0 },
  systemInfo: { osVersion: '', appVersion: '', lastUpdated: new Date().toISOString() },
  timestamps: { createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  commands: { type: 'ssh', restart: { enabled: false } },
  upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
});

describe('uiSlice', () => {
  const initialState = {
    sidebarOpen: true,
    darkMode: true,
    environmentCreateDialogOpen: false,
    environmentEditDialogOpen: false,
    selectedEnvironment: null,
    confirmDialogOpen: false,
    confirmDialog: null,
  };

  it('returns initial state', () => {
    const state = reducer(undefined, { type: '@@INIT' });
    expect(state.sidebarOpen).toBe(true);
    expect(state.darkMode).toBe(true);
  });

  it('toggleSidebar flips sidebarOpen', () => {
    let state = reducer(initialState, toggleSidebar());
    expect(state.sidebarOpen).toBe(false);
    state = reducer(state, toggleSidebar());
    expect(state.sidebarOpen).toBe(true);
  });

  it('setSidebarOpen sets explicitly', () => {
    const state = reducer(initialState, setSidebarOpen(false));
    expect(state.sidebarOpen).toBe(false);
  });

  it('toggleDarkMode flips darkMode', () => {
    let state = reducer(initialState, toggleDarkMode());
    expect(state.darkMode).toBe(false);
    state = reducer(state, toggleDarkMode());
    expect(state.darkMode).toBe(true);
  });

  it('setEnvironmentCreateDialogOpen sets flag', () => {
    const state = reducer(initialState, setEnvironmentCreateDialogOpen(true));
    expect(state.environmentCreateDialogOpen).toBe(true);
  });

  it('setEnvironmentEditDialogOpen(true) sets flag', () => {
    const state = reducer(initialState, setEnvironmentEditDialogOpen(true));
    expect(state.environmentEditDialogOpen).toBe(true);
  });

  it('setEnvironmentEditDialogOpen(false) clears selectedEnvironment', () => {
    const env = makeEnv();
    const stateWithEnv = { ...initialState, selectedEnvironment: env, environmentEditDialogOpen: true };
    const state = reducer(stateWithEnv, setEnvironmentEditDialogOpen(false));
    expect(state.selectedEnvironment).toBeNull();
    expect(state.environmentEditDialogOpen).toBe(false);
  });

  it('setSelectedEnvironment sets and clears', () => {
    const env = makeEnv();
    let state = reducer(initialState, setSelectedEnvironment(env));
    expect(state.selectedEnvironment?.id).toBe('env-1');
    state = reducer(state, setSelectedEnvironment(null));
    expect(state.selectedEnvironment).toBeNull();
  });

  it('openConfirmDialog sets confirmDialogOpen and stores dialog config', () => {
    const onConfirm = vi.fn();
    const state = reducer(initialState, openConfirmDialog({
      title: 'Delete?',
      message: 'Are you sure?',
      onConfirm,
    }));
    expect(state.confirmDialogOpen).toBe(true);
    expect(state.confirmDialog?.title).toBe('Delete?');
    expect(state.confirmDialog?.message).toBe('Are you sure?');
    expect(state.confirmDialog?.onConfirm).toBe(onConfirm);
  });

  it('closeConfirmDialog resets dialog state', () => {
    const onConfirm = vi.fn();
    let state = reducer(initialState, openConfirmDialog({ title: 'T', message: 'M', onConfirm }));
    state = reducer(state, closeConfirmDialog());
    expect(state.confirmDialogOpen).toBe(false);
    expect(state.confirmDialog).toBeNull();
  });
});
