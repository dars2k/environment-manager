import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Environment } from '@/types/environment';

interface EnvironmentState {
  environments: Environment[];
  selectedEnvironment: Environment | null;
  loading: boolean;
  error: string | null;
}

const initialState: EnvironmentState = {
  environments: [],
  selectedEnvironment: null,
  loading: false,
  error: null,
};

const environmentSlice = createSlice({
  name: 'environments',
  initialState,
  reducers: {
    setEnvironments: (state, action: PayloadAction<Environment[]>) => {
      state.environments = action.payload;
      state.error = null;
    },
    addEnvironment: (state, action: PayloadAction<Environment>) => {
      state.environments.push(action.payload);
    },
    updateEnvironment: (state, action: PayloadAction<Environment>) => {
      const index = state.environments.findIndex(env => env.id === action.payload.id);
      if (index !== -1) {
        state.environments[index] = action.payload;
      }
      if (state.selectedEnvironment?.id === action.payload.id) {
        state.selectedEnvironment = action.payload;
      }
    },
    removeEnvironment: (state, action: PayloadAction<string>) => {
      state.environments = state.environments.filter(env => env.id !== action.payload);
      if (state.selectedEnvironment?.id === action.payload) {
        state.selectedEnvironment = null;
      }
    },
    setSelectedEnvironment: (state, action: PayloadAction<Environment | null>) => {
      state.selectedEnvironment = action.payload;
    },
    updateEnvironmentStatus: (state, action: PayloadAction<{ id: string; status: any }>) => {
      const env = state.environments.find(e => e.id === action.payload.id);
      if (env) {
        env.status = action.payload.status;
      }
      if (state.selectedEnvironment?.id === action.payload.id) {
        state.selectedEnvironment.status = action.payload.status;
      }
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
  },
});

export const {
  setEnvironments,
  addEnvironment,
  updateEnvironment,
  removeEnvironment,
  setSelectedEnvironment,
  updateEnvironmentStatus,
  setLoading,
  setError,
} = environmentSlice.actions;

export default environmentSlice.reducer;
