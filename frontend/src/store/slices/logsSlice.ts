import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface LogsState {
  unreadCount: number;
  lastViewedTimestamp: string | null;
}

const initialState: LogsState = {
  unreadCount: 0,
  lastViewedTimestamp: localStorage.getItem('lastViewedLogsTimestamp'),
};

const logsSlice = createSlice({
  name: 'logs',
  initialState,
  reducers: {
    setUnreadCount: (state, action: PayloadAction<number>) => {
      state.unreadCount = action.payload;
    },
    incrementUnreadCount: (state, action: PayloadAction<number>) => {
      state.unreadCount += action.payload;
    },
    clearUnreadCount: (state) => {
      state.unreadCount = 0;
      state.lastViewedTimestamp = new Date().toISOString();
      localStorage.setItem('lastViewedLogsTimestamp', state.lastViewedTimestamp);
    },
    setLastViewedTimestamp: (state, action: PayloadAction<string>) => {
      state.lastViewedTimestamp = action.payload;
      localStorage.setItem('lastViewedLogsTimestamp', action.payload);
    },
  },
});

export const {
  setUnreadCount,
  incrementUnreadCount,
  clearUnreadCount,
  setLastViewedTimestamp,
} = logsSlice.actions;

export default logsSlice.reducer;
