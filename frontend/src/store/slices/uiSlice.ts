import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Environment } from '@/types/environment';

interface UiState {
  sidebarOpen: boolean;
  darkMode: boolean;
  environmentCreateDialogOpen: boolean;
  environmentEditDialogOpen: boolean;
  selectedEnvironment: Environment | null;
  confirmDialogOpen: boolean;
  confirmDialog: {
    title: string;
    message: string;
    onConfirm: () => void;
  } | null;
}

const initialState: UiState = {
  sidebarOpen: true,
  darkMode: true,
  environmentCreateDialogOpen: false,
  environmentEditDialogOpen: false,
  selectedEnvironment: null,
  confirmDialogOpen: false,
  confirmDialog: null,
};

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    toggleSidebar: (state) => {
      state.sidebarOpen = !state.sidebarOpen;
    },
    setSidebarOpen: (state, action: PayloadAction<boolean>) => {
      state.sidebarOpen = action.payload;
    },
    toggleDarkMode: (state) => {
      state.darkMode = !state.darkMode;
    },
    setEnvironmentCreateDialogOpen: (state, action: PayloadAction<boolean>) => {
      state.environmentCreateDialogOpen = action.payload;
    },
    setEnvironmentEditDialogOpen: (state, action: PayloadAction<boolean>) => {
      state.environmentEditDialogOpen = action.payload;
      if (!action.payload) {
        state.selectedEnvironment = null;
      }
    },
    setSelectedEnvironment: (state, action: PayloadAction<Environment | null>) => {
      state.selectedEnvironment = action.payload;
    },
    openConfirmDialog: (state, action: PayloadAction<{
      title: string;
      message: string;
      onConfirm: () => void;
    }>) => {
      state.confirmDialogOpen = true;
      state.confirmDialog = action.payload;
    },
    closeConfirmDialog: (state) => {
      state.confirmDialogOpen = false;
      state.confirmDialog = null;
    },
  },
});

export const {
  toggleSidebar,
  setSidebarOpen,
  toggleDarkMode,
  setEnvironmentCreateDialogOpen,
  setEnvironmentEditDialogOpen,
  setSelectedEnvironment,
  openConfirmDialog,
  closeConfirmDialog,
} = uiSlice.actions;

export default uiSlice.reducer;
