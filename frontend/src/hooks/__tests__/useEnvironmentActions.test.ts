import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useEnvironmentActions } from '../useEnvironmentActions';
import { environmentApi } from '@/api/environments';
import { createTestStore } from '@/test/test-utils';
import { Provider } from 'react-redux';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('@/api/environments');
const mockedApi = vi.mocked(environmentApi);

function createWrapper() {
  const store = createTestStore();
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return {
    store,
    wrapper: ({ children }: { children: React.ReactNode }) =>
      React.createElement(
        Provider,
        { store },
        React.createElement(QueryClientProvider, { client: queryClient, children })
      ),
  };
}

describe('useEnvironmentActions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns expected functions and state', () => {
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });
    expect(typeof result.current.restart).toBe('function');
    expect(typeof result.current.upgrade).toBe('function');
    expect(typeof result.current.deleteEnvironment).toBe('function');
    expect(result.current.isRestarting).toBe(false);
    expect(result.current.isUpgrading).toBe(false);
    expect(result.current.isDeleting).toBe(false);
  });

  it('opens confirm dialog on force restart', () => {
    const { store, wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });
    act(() => {
      result.current.restart('env-1', true);
    });
    const state = store.getState().ui;
    expect(state.confirmDialogOpen).toBe(true);
    expect(state.confirmDialog?.title).toMatch(/force restart/i);
  });

  it('calls restart API directly on non-force restart', async () => {
    mockedApi.restart = vi.fn().mockResolvedValue({});
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });
    await act(async () => {
      result.current.restart('env-1', false);
    });
    expect(mockedApi.restart).toHaveBeenCalledWith('env-1', false);
  });

  it('opens confirm dialog on upgrade', () => {
    const { store, wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });
    act(() => {
      result.current.upgrade('env-1', '2.0.0');
    });
    const state = store.getState().ui;
    expect(state.confirmDialogOpen).toBe(true);
    expect(state.confirmDialog?.title).toMatch(/upgrade/i);
  });

  it('opens confirm dialog on delete', () => {
    const { store, wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });
    act(() => {
      result.current.deleteEnvironment('env-1');
    });
    const state = store.getState().ui;
    expect(state.confirmDialogOpen).toBe(true);
    expect(state.confirmDialog?.title).toMatch(/delete/i);
  });

  it('dispatches showSuccess on successful restart', async () => {
    mockedApi.restart = vi.fn().mockResolvedValue({ operationId: 'op-1', status: 'pending' });
    const { store, wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });
    await act(async () => {
      result.current.restart('env-1', false);
    });
    await act(async () => {});
    const notifications = store.getState().notifications.notifications;
    expect(notifications.some((n: any) => n.type === 'success')).toBe(true);
  });

  it('dispatches showError on failed restart', async () => {
    mockedApi.restart = vi.fn().mockRejectedValue(new Error('restart failed'));
    const { store, wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });
    await act(async () => {
      result.current.restart('env-1', false);
    });
    await act(async () => {});
    const notifications = store.getState().notifications.notifications;
    expect(notifications.some((n: any) => n.type === 'error')).toBe(true);
  });

  it('dispatches showSuccess on successful upgrade (via confirm callback)', async () => {
    mockedApi.upgrade = vi.fn().mockResolvedValue({ operationId: 'op-2', status: 'pending' });
    const { store, wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });

    // Trigger upgrade which opens confirm dialog
    act(() => {
      result.current.upgrade('env-1', '2.0.0');
    });

    // Get the onConfirm from the dialog state and call it
    const dialogState = store.getState().ui.confirmDialog;
    expect(dialogState?.onConfirm).toBeDefined();

    await act(async () => {
      dialogState?.onConfirm?.();
    });
    await act(async () => {});

    const notifications = store.getState().notifications.notifications;
    expect(notifications.some((n: any) => n.type === 'success')).toBe(true);
  });

  it('dispatches showError on failed upgrade', async () => {
    mockedApi.upgrade = vi.fn().mockRejectedValue(new Error('upgrade failed'));
    const { store, wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });

    act(() => {
      result.current.upgrade('env-1', '2.0.0');
    });

    const dialogState = store.getState().ui.confirmDialog;
    await act(async () => {
      dialogState?.onConfirm?.();
    });
    await act(async () => {});

    const notifications = store.getState().notifications.notifications;
    expect(notifications.some((n: any) => n.type === 'error')).toBe(true);
  });

  it('dispatches removeEnvironment and showSuccess on successful delete', async () => {
    mockedApi.delete = vi.fn().mockResolvedValue(undefined);
    const { store, wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });

    act(() => {
      result.current.deleteEnvironment('env-1');
    });

    const dialogState = store.getState().ui.confirmDialog;
    await act(async () => {
      dialogState?.onConfirm?.();
    });
    await act(async () => {});

    const notifications = store.getState().notifications.notifications;
    expect(notifications.some((n: any) => n.type === 'success')).toBe(true);
  });

  it('dispatches showError on failed delete', async () => {
    mockedApi.delete = vi.fn().mockRejectedValue(new Error('delete failed'));
    const { store, wrapper } = createWrapper();
    const { result } = renderHook(() => useEnvironmentActions(), { wrapper });

    act(() => {
      result.current.deleteEnvironment('env-1');
    });

    const dialogState = store.getState().ui.confirmDialog;
    await act(async () => {
      dialogState?.onConfirm?.();
    });
    await act(async () => {});

    const notifications = store.getState().notifications.notifications;
    expect(notifications.some((n: any) => n.type === 'error')).toBe(true);
  });
});
