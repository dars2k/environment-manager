import { describe, it, expect } from 'vitest';
import reducer, {
  addNotification,
  removeNotification,
  clearNotifications,
  showSuccess,
  showError,
  showWarning,
  showInfo,
} from '../notificationSlice';

describe('notificationSlice', () => {
  const initialState = { notifications: [] };

  it('returns initial state', () => {
    expect(reducer(undefined, { type: '@@INIT' })).toEqual(initialState);
  });

  it('addNotification appends notification with generated id', () => {
    const state = reducer(initialState, addNotification({ type: 'success', message: 'Done!' }));
    expect(state.notifications).toHaveLength(1);
    expect(state.notifications[0].type).toBe('success');
    expect(state.notifications[0].message).toBe('Done!');
    expect(state.notifications[0].id).toBeDefined();
  });

  it('addNotification with optional duration', () => {
    const state = reducer(initialState, addNotification({ type: 'error', message: 'Err', duration: 5000 }));
    expect(state.notifications[0].duration).toBe(5000);
  });

  it('removeNotification filters by id', () => {
    // Build state with two notifications by starting from initial each time
    const state1 = reducer(initialState, addNotification({ type: 'info', message: 'msg1' }));
    const id1 = state1.notifications[0].id;
    const state2 = { notifications: [...state1.notifications, { id: 'fixed-id-2', type: 'warning' as const, message: 'msg2' }] };
    const state3 = reducer(state2, removeNotification(id1));
    expect(state3.notifications).toHaveLength(1);
    expect(state3.notifications[0].message).toBe('msg2');
  });

  it('clearNotifications empties the list', () => {
    let state = reducer(initialState, addNotification({ type: 'success', message: 'a' }));
    state = reducer(state, addNotification({ type: 'success', message: 'b' }));
    state = reducer(state, clearNotifications());
    expect(state.notifications).toHaveLength(0);
  });

  it('showSuccess creates correct action', () => {
    const action = showSuccess('Saved!');
    expect(action.type).toBe('notifications/addNotification');
    expect(action.payload.type).toBe('success');
    expect(action.payload.message).toBe('Saved!');
  });

  it('showSuccess with duration', () => {
    const action = showSuccess('Done', 3000);
    expect(action.payload.duration).toBe(3000);
  });

  it('showError creates correct action', () => {
    const action = showError('Failed!');
    expect(action.payload.type).toBe('error');
    expect(action.payload.message).toBe('Failed!');
  });

  it('showWarning creates correct action', () => {
    const action = showWarning('Careful!');
    expect(action.payload.type).toBe('warning');
  });

  it('showInfo creates correct action', () => {
    const action = showInfo('FYI');
    expect(action.payload.type).toBe('info');
  });
});
