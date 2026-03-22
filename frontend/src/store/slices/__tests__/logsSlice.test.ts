import { describe, it, expect, vi, beforeEach } from 'vitest';
import reducer, {
  setUnreadCount,
  incrementUnreadCount,
  clearUnreadCount,
  setLastViewedTimestamp,
} from '../logsSlice';

describe('logsSlice', () => {
  beforeEach(() => {
    vi.spyOn(localStorage, 'getItem').mockReturnValue(null);
    vi.spyOn(localStorage, 'setItem').mockImplementation(() => {});
  });

  it('returns initial state with null lastViewedTimestamp', () => {
    const state = reducer(undefined, { type: '@@INIT' });
    expect(state.unreadCount).toBe(0);
    // lastViewedTimestamp comes from localStorage which is mocked to null
    expect(state.lastViewedTimestamp).toBeNull();
  });

  it('setUnreadCount sets count', () => {
    const state = reducer({ unreadCount: 0, lastViewedTimestamp: null }, setUnreadCount(5));
    expect(state.unreadCount).toBe(5);
  });

  it('incrementUnreadCount adds to count', () => {
    const state = reducer({ unreadCount: 3, lastViewedTimestamp: null }, incrementUnreadCount(4));
    expect(state.unreadCount).toBe(7);
  });

  it('clearUnreadCount resets count and updates timestamp', () => {
    const state = reducer({ unreadCount: 10, lastViewedTimestamp: null }, clearUnreadCount());
    expect(state.unreadCount).toBe(0);
    expect(state.lastViewedTimestamp).not.toBeNull();
    expect(localStorage.setItem).toHaveBeenCalledWith('lastViewedLogsTimestamp', expect.any(String));
  });

  it('setLastViewedTimestamp sets timestamp and persists', () => {
    const ts = '2024-01-01T00:00:00.000Z';
    const state = reducer({ unreadCount: 0, lastViewedTimestamp: null }, setLastViewedTimestamp(ts));
    expect(state.lastViewedTimestamp).toBe(ts);
    expect(localStorage.setItem).toHaveBeenCalledWith('lastViewedLogsTimestamp', ts);
  });
});
