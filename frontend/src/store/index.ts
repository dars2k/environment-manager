import { configureStore } from '@reduxjs/toolkit';
import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';

import environmentReducer from './slices/environmentSlice';
import uiReducer from './slices/uiSlice';
import notificationReducer from './slices/notificationSlice';
import logsReducer from './slices/logsSlice';

export const store = configureStore({
  reducer: {
    environments: environmentReducer,
    ui: uiReducer,
    notifications: notificationReducer,
    logs: logsReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

// Typed hooks
export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;
