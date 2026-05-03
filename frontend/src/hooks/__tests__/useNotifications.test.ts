import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useNotifications } from '../useNotifications';
import { createTestStore } from '@/test/test-utils';
import { addNotification } from '@/store/slices/notificationSlice';
import React from 'react';
import { Provider } from 'react-redux';
import { SnackbarProvider } from 'notistack';

const mockEnqueueSnackbar = vi.fn();

vi.mock('notistack', () => ({
  useSnackbar: () => ({
    enqueueSnackbar: mockEnqueueSnackbar,
    closeSnackbar: vi.fn(),
  }),
  SnackbarProvider: ({ children }: { children: React.ReactNode }) => children,
}));

function createWrapper(store: ReturnType<typeof createTestStore>) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(
      Provider,
      { store },
      React.createElement(SnackbarProvider, null, children)
    );
}

describe('useNotifications hook', () => {
  let store: ReturnType<typeof createTestStore>;

  beforeEach(() => {
    vi.clearAllMocks();
    store = createTestStore();
  });

  it('renders without crashing', () => {
    const { result } = renderHook(() => useNotifications(), {
      wrapper: createWrapper(store),
    });
    // hook returns undefined (it's a side-effect hook)
    expect(result.current).toBeUndefined();
  });

  it('calls enqueueSnackbar when a notification is added', () => {
    const { rerender } = renderHook(() => useNotifications(), {
      wrapper: createWrapper(store),
    });

    store.dispatch(
      addNotification({
        message: 'Test notification',
        type: 'success',
      })
    );

    rerender();

    expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
      'Test notification',
      expect.objectContaining({
        variant: 'success',
      })
    );
  });

  it('does not call enqueueSnackbar for the same notification twice', () => {
    const { rerender } = renderHook(() => useNotifications(), {
      wrapper: createWrapper(store),
    });

    store.dispatch(
      addNotification({
        message: 'Test notification 2',
        type: 'info',
      })
    );

    rerender();
    rerender();

    expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1);
  });

  it('uses default duration of 5000 when not specified', () => {
    const { rerender } = renderHook(() => useNotifications(), {
      wrapper: createWrapper(store),
    });

    store.dispatch(
      addNotification({
        message: 'Test notification 3',
        type: 'warning',
      })
    );

    rerender();

    expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
      'Test notification 3',
      expect.objectContaining({
        autoHideDuration: 5000,
      })
    );
  });

  it('uses custom duration when specified', () => {
    const { rerender } = renderHook(() => useNotifications(), {
      wrapper: createWrapper(store),
    });

    store.dispatch(
      addNotification({
        message: 'Test notification 4',
        type: 'error',
        duration: 3000,
      })
    );

    rerender();

    expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
      'Test notification 4',
      expect.objectContaining({
        autoHideDuration: 3000,
      })
    );
  });

  it('skips already-displayed notifications when new ones are added', () => {
    const { rerender } = renderHook(() => useNotifications(), {
      wrapper: createWrapper(store),
    });

    // First notification
    store.dispatch(addNotification({ message: 'First', type: 'success' }));
    rerender();
    expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1);

    // Add a second notification — effect reruns with new array reference,
    // the first notification hits the "already displayed" early-return path.
    store.dispatch(addNotification({ message: 'Second', type: 'info' }));
    rerender();

    // enqueueSnackbar called only once more (for 'Second'), not again for 'First'
    expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(2);
    expect(mockEnqueueSnackbar).toHaveBeenLastCalledWith(
      'Second',
      expect.objectContaining({ variant: 'info' })
    );
  });

  it('onExited callback dispatches removeNotification and cleans up the displayed set', () => {
    const { rerender } = renderHook(() => useNotifications(), {
      wrapper: createWrapper(store),
    });

    store.dispatch(addNotification({ message: 'Exiting', type: 'success' }));
    rerender();

    expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1);

    // Extract the onExited callback from the enqueueSnackbar call and invoke it
    const callArgs = mockEnqueueSnackbar.mock.calls[0][1] as { onExited?: () => void };
    expect(callArgs.onExited).toBeDefined();
    callArgs.onExited?.();

    // After onExited, the notification should be removed from the store
    const state = store.getState();
    expect(state.notifications.notifications).toHaveLength(0);
  });

  it('allows the same notification id to be displayed again after onExited cleans it up', () => {
    const { rerender } = renderHook(() => useNotifications(), {
      wrapper: createWrapper(store),
    });

    store.dispatch(addNotification({ message: 'Re-show', type: 'warning' }));
    rerender();
    expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1);

    // Invoke onExited to clean up the displayed-set entry
    const callArgs = mockEnqueueSnackbar.mock.calls[0][1] as { onExited?: () => void };
    callArgs.onExited?.();

    // Re-add the same message — in practice this would get a new ID via addNotification,
    // so a fresh notification can be displayed.
    store.dispatch(addNotification({ message: 'Re-show', type: 'warning' }));
    rerender();

    // A new notification (new id) should be enqueued
    expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(2);
  });
});
