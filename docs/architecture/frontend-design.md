# Frontend Architecture Design

## React Project Structure

The frontend follows a feature-based architecture with Redux Toolkit for state management:

```
frontend/
├── public/
│   └── (static assets)
├── src/
│   ├── api/                   # API service layer
│   │   ├── environments.ts
│   │   └── users.ts
│   ├── components/            # Reusable UI components
│   │   ├── common/           # Generic components
│   │   │   └── ConfirmDialog.tsx
│   │   ├── environments/     # Environment-specific components
│   │   │   ├── CreateEnvironmentDialog.tsx
│   │   │   ├── EditEnvironmentDialog.tsx
│   │   │   ├── EnvironmentCard.tsx
│   │   │   ├── EnvironmentForm.tsx
│   │   │   ├── EnvironmentLogs.tsx
│   │   │   ├── HttpRequestConfig.tsx
│   │   │   ├── UpgradeEnvironmentDialog.tsx
│   │   │   └── index.ts
│   │   └── layout/           # Layout components
│   ├── contexts/             # React contexts
│   │   └── WebSocketContext.tsx
│   ├── hooks/                # Custom React hooks
│   │   └── useEnvironmentActions.ts
│   ├── layouts/              # Page layouts
│   │   └── MainLayout.tsx
│   ├── pages/                # Page components
│   │   ├── CreateEnvironment.tsx
│   │   ├── Dashboard.tsx
│   │   ├── EditEnvironment.tsx
│   │   ├── EnvironmentDetails.tsx
│   │   ├── Login.tsx
│   │   ├── Logs.tsx
│   │   ├── NotFound.tsx
│   │   └── Users.tsx
│   ├── routes/               # Routing configuration
│   │   └── index.tsx
│   ├── store/                # Redux store configuration
│   │   ├── index.ts
│   │   └── slices/          # Redux slices
│   │       ├── environmentSlice.ts
│   │       ├── logsSlice.ts
│   │       ├── notificationSlice.ts
│   │       └── uiSlice.ts
│   ├── theme/                # MUI theme configuration
│   │   └── index.ts
│   ├── types/                # TypeScript type definitions
│   │   └── environment.ts
│   ├── utils/                # Utility functions
│   ├── App.tsx
│   └── main.tsx
├── index.html
├── vite.config.ts
├── tsconfig.json
├── package.json
└── Dockerfile
```

## Core Technologies

- **Framework**: React 18 with TypeScript
- **State Management**: Redux Toolkit
- **UI Library**: Material-UI (MUI) v5
- **Routing**: React Router v6
- **Build Tool**: Vite
- **HTTP Client**: Axios
- **WebSocket**: Native WebSocket API with React Context

## State Management with Redux Toolkit

### Store Configuration

```typescript
// src/store/index.ts
import { configureStore } from '@reduxjs/toolkit';
import environmentReducer from './slices/environmentSlice';
import logsReducer from './slices/logsSlice';
import uiReducer from './slices/uiSlice';
import notificationReducer from './slices/notificationSlice';

export const store = configureStore({
  reducer: {
    environments: environmentReducer,
    logs: logsReducer,
    ui: uiReducer,
    notifications: notificationReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

### Environment Slice

```typescript
// src/store/slices/environmentSlice.ts
import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { environmentsApi } from '@/api/environments';
import { Environment } from '@/types/environment';

interface EnvironmentState {
  environments: Environment[];
  selectedEnvironment: Environment | null;
  loading: boolean;
  error: string | null;
}

export const fetchEnvironments = createAsyncThunk(
  'environments/fetchAll',
  async () => {
    const response = await environmentsApi.getAll();
    return response.data;
  }
);

const environmentSlice = createSlice({
  name: 'environments',
  initialState: {
    environments: [],
    selectedEnvironment: null,
    loading: false,
    error: null,
  } as EnvironmentState,
  reducers: {
    selectEnvironment: (state, action: PayloadAction<string>) => {
      state.selectedEnvironment = state.environments.find(
        env => env.id === action.payload
      ) || null;
    },
    updateEnvironmentStatus: (state, action: PayloadAction<{
      id: string;
      status: Environment['status'];
    }>) => {
      const env = state.environments.find(e => e.id === action.payload.id);
      if (env) {
        env.status = action.payload.status;
      }
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchEnvironments.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchEnvironments.fulfilled, (state, action) => {
        state.loading = false;
        state.environments = action.payload;
      })
      .addCase(fetchEnvironments.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch environments';
      });
  },
});

export const { selectEnvironment, updateEnvironmentStatus } = environmentSlice.actions;
export default environmentSlice.reducer;
```

## Core Components

### Main Layout

```tsx
// src/layouts/MainLayout.tsx
import React from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  AppBar,
  Box,
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Typography,
  IconButton,
  Avatar,
  Menu,
  MenuItem,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  ViewList as LogsIcon,
  People as UsersIcon,
  ExitToApp as LogoutIcon,
} from '@mui/icons-material';

const DRAWER_WIDTH = 240;

export const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);

  const menuItems = [
    { path: '/', label: 'Dashboard', icon: <DashboardIcon /> },
    { path: '/logs', label: 'Logs', icon: <LogsIcon /> },
    { path: '/users', label: 'Users', icon: <UsersIcon /> },
  ];

  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar position="fixed">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            Environment Manager
          </Typography>
          <IconButton onClick={(e) => setAnchorEl(e.currentTarget)}>
            <Avatar />
          </IconButton>
          <Menu
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={() => setAnchorEl(null)}
          >
            <MenuItem onClick={() => navigate('/profile')}>Profile</MenuItem>
            <MenuItem onClick={() => navigate('/settings')}>Settings</MenuItem>
            <MenuItem onClick={handleLogout}>
              <LogoutIcon sx={{ mr: 1 }} /> Logout
            </MenuItem>
          </Menu>
        </Toolbar>
      </AppBar>
      
      <Drawer
        variant="permanent"
        sx={{
          width: DRAWER_WIDTH,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: DRAWER_WIDTH,
            boxSizing: 'border-box',
            top: 64,
          },
        }}
      >
        <List>
          {menuItems.map((item) => (
            <ListItem
              button
              key={item.path}
              selected={location.pathname === item.path}
              onClick={() => navigate(item.path)}
            >
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.label} />
            </ListItem>
          ))}
        </List>
      </Drawer>
      
      <Box component="main" sx={{ flexGrow: 1, p: 3, mt: 8 }}>
        <Outlet />
      </Box>
    </Box>
  );
};
```

### Dashboard Page

```tsx
// src/pages/Dashboard.tsx
import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { Grid, Typography, Box, Fab } from '@mui/material';
import { Add as AddIcon } from '@mui/icons-material';
import { EnvironmentCard } from '@/components/environments';
import { fetchEnvironments } from '@/store/slices/environmentSlice';
import { RootState, AppDispatch } from '@/store';

export const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();
  const { environments, loading } = useSelector((state: RootState) => state.environments);

  useEffect(() => {
    dispatch(fetchEnvironments());
  }, [dispatch]);

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Environments
      </Typography>
      
      <Grid container spacing={3}>
        {environments.map((environment) => (
          <Grid item xs={12} sm={6} md={4} key={environment.id}>
            <EnvironmentCard
              environment={environment}
              onClick={() => navigate(`/environments/${environment.id}`)}
            />
          </Grid>
        ))}
      </Grid>
      
      <Fab
        color="primary"
        aria-label="add"
        sx={{
          position: 'fixed',
          bottom: 24,
          right: 24,
        }}
        onClick={() => navigate('/environments/create')}
      >
        <AddIcon />
      </Fab>
    </Box>
  );
};
```

### Environment Form Component

```tsx
// src/components/environments/EnvironmentForm.tsx
import React from 'react';
import { useForm, Controller } from 'react-hook-form';
import {
  TextField,
  Switch,
  FormControlLabel,
  Box,
  Typography,
  Tabs,
  Tab,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
} from '@mui/material';
import { Environment } from '@/types/environment';
import { HttpRequestConfig } from './HttpRequestConfig';

interface EnvironmentFormProps {
  environment?: Environment;
  onSubmit: (data: Partial<Environment>) => void;
}

export const EnvironmentForm: React.FC<EnvironmentFormProps> = ({ 
  environment, 
  onSubmit 
}) => {
  const { control, handleSubmit, watch } = useForm({
    defaultValues: environment || {
      name: '',
      description: '',
      environmentURL: '',
      target: { host: '', port: 22 },
      credentials: { type: 'key', username: '' },
      healthCheck: { enabled: true, interval: 30, timeout: 5 },
      commands: { type: 'ssh', restart: { command: '' } },
    }
  });
  
  const [tabValue, setTabValue] = React.useState(0);
  const commandType = watch('commands.type');

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Tabs value={tabValue} onChange={(_, v) => setTabValue(v)}>
        <Tab label="Basic Info" />
        <Tab label="Connection" />
        <Tab label="Health Check" />
        <Tab label="Commands" />
        <Tab label="Upgrade Config" />
      </Tabs>
      
      {tabValue === 0 && (
        <Box sx={{ mt: 3 }}>
          <Controller
            name="name"
            control={control}
            rules={{ required: 'Name is required' }}
            render={({ field, fieldState }) => (
              <TextField
                {...field}
                label="Name"
                fullWidth
                margin="normal"
                error={!!fieldState.error}
                helperText={fieldState.error?.message}
              />
            )}
          />
          
          <Controller
            name="description"
            control={control}
            render={({ field }) => (
              <TextField
                {...field}
                label="Description"
                fullWidth
                margin="normal"
                multiline
                rows={3}
              />
            )}
          />
          
          <Controller
            name="environmentURL"
            control={control}
            render={({ field }) => (
              <TextField
                {...field}
                label="Environment URL"
                fullWidth
                margin="normal"
                placeholder="https://example.com"
              />
            )}
          />
        </Box>
      )}
      
      {tabValue === 3 && (
        <Box sx={{ mt: 3 }}>
          <Controller
            name="commands.type"
            control={control}
            render={({ field }) => (
              <FormControl fullWidth margin="normal">
                <InputLabel>Command Type</InputLabel>
                <Select {...field} label="Command Type">
                  <MenuItem value="ssh">SSH</MenuItem>
                  <MenuItem value="http">HTTP</MenuItem>
                </Select>
              </FormControl>
            )}
          />
          
          {commandType === 'ssh' ? (
            <Controller
              name="commands.restart.command"
              control={control}
              render={({ field }) => (
                <TextField
                  {...field}
                  label="Restart Command"
                  fullWidth
                  margin="normal"
                  placeholder="sudo systemctl restart myservice"
                />
              )}
            />
          ) : (
            <HttpRequestConfig
              control={control}
              prefix="commands.restart"
            />
          )}
        </Box>
      )}
    </form>
  );
};
```

## API Service Layer

```typescript
// src/api/environments.ts
import axios from 'axios';
import { Environment } from '@/types/environment';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const environmentsApi = {
  getAll: () => apiClient.get<Environment[]>('/environments'),
  
  getById: (id: string) => apiClient.get<Environment>(`/environments/${id}`),
  
  create: (data: Partial<Environment>) => 
    apiClient.post<Environment>('/environments', data),
  
  update: (id: string, data: Partial<Environment>) => 
    apiClient.put<Environment>(`/environments/${id}`, data),
  
  delete: (id: string) => apiClient.delete(`/environments/${id}`),
  
  restart: (id: string, data?: { force?: boolean }) => 
    apiClient.post(`/environments/${id}/restart`, data),
  
  checkHealth: (id: string) => 
    apiClient.post(`/environments/${id}/check-health`),
  
  getVersions: (id: string) => 
    apiClient.get(`/environments/${id}/versions`),
  
  upgrade: (id: string, data: { version: string }) => 
    apiClient.post(`/environments/${id}/upgrade`, data),
  
  getLogs: (id: string, params?: any) => 
    apiClient.get(`/environments/${id}/logs`, { params }),
};
```

## WebSocket Integration

```tsx
// src/contexts/WebSocketContext.tsx
import React, { createContext, useContext, useEffect, useRef } from 'react';
import { useDispatch } from 'react-redux';
import { updateEnvironmentStatus } from '@/store/slices/environmentSlice';
import { addNotification } from '@/store/slices/notificationSlice';

interface WebSocketContextValue {
  isConnected: boolean;
  subscribe: (environmentIds: string[]) => void;
  unsubscribe: (environmentIds: string[]) => void;
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null);

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({ 
  children 
}) => {
  const dispatch = useDispatch();
  const ws = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = React.useState(false);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) return;

    const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost:8080'}/ws`;
    ws.current = new WebSocket(wsUrl);

    ws.current.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
    };

    ws.current.onmessage = (event) => {
      const message = JSON.parse(event.data);
      
      switch (message.type) {
        case 'status_update':
          dispatch(updateEnvironmentStatus({
            id: message.payload.environmentId,
            status: message.payload.status,
          }));
          break;
          
        case 'operation_update':
          dispatch(addNotification({
            type: 'info',
            message: `Operation ${message.payload.operationId} ${message.payload.status}`,
          }));
          break;
      }
    };

    ws.current.onclose = () => {
      console.log('WebSocket disconnected');
      setIsConnected(false);
    };

    return () => {
      ws.current?.close();
    };
  }, [dispatch]);

  const subscribe = (environmentIds: string[]) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({
        type: 'subscribe',
        payload: { environments: environmentIds },
      }));
    }
  };

  const unsubscribe = (environmentIds: string[]) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({
        type: 'unsubscribe',
        payload: { environments: environmentIds },
      }));
    }
  };

  return (
    <WebSocketContext.Provider value={{ isConnected, subscribe, unsubscribe }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within WebSocketProvider');
  }
  return context;
};
```

## Custom Hooks

```typescript
// src/hooks/useEnvironmentActions.ts
import { useState } from 'react';
import { useDispatch } from 'react-redux';
import { environmentsApi } from '@/api/environments';
import { addNotification } from '@/store/slices/notificationSlice';
import { fetchEnvironments } from '@/store/slices/environmentSlice';

export const useEnvironmentActions = (environmentId: string) => {
  const dispatch = useDispatch();
  const [loading, setLoading] = useState(false);

  const restart = async (force = false) => {
    setLoading(true);
    try {
      await environmentsApi.restart(environmentId, { force });
      dispatch(addNotification({
        type: 'success',
        message: 'Environment restart initiated',
      }));
    } catch (error) {
      dispatch(addNotification({
        type: 'error',
        message: 'Failed to restart environment',
      }));
    } finally {
      setLoading(false);
    }
  };

  const checkHealth = async () => {
    setLoading(true);
    try {
      await environmentsApi.checkHealth(environmentId);
      dispatch(fetchEnvironments());
      dispatch(addNotification({
        type: 'success',
        message: 'Health check completed',
      }));
    } catch (error) {
      dispatch(addNotification({
        type: 'error',
        message: 'Health check failed',
      }));
    } finally {
      setLoading(false);
    }
  };

  const upgrade = async (version: string) => {
    setLoading(true);
    try {
      await environmentsApi.upgrade(environmentId, { version });
      dispatch(addNotification({
        type: 'success',
        message: 'Environment upgrade initiated',
      }));
    } catch (error) {
      dispatch(addNotification({
        type: 'error',
        message: 'Failed to upgrade environment',
      }));
    } finally {
      setLoading(false);
    }
  };

  return { restart, checkHealth, upgrade, loading };
};
```

## Theme Configuration

```typescript
// src/theme/index.ts
import { createTheme } from '@mui/material/styles';

export const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#3b82f6',
    },
    secondary: {
      main: '#8b5cf6',
    },
    success: {
      main: '#10b981',
    },
    error: {
      main: '#ef4444',
    },
    warning: {
      main: '#f59e0b',
    },
    background: {
      default: '#0f172a',
      paper: '#1e293b',
    },
  },
  typography: {
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: 8,
        },
      },
    },
  },
});
```

## Routing Configuration

```tsx
// src/routes/index.tsx
import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { MainLayout } from '@/layouts/MainLayout';
import { Login } from '@/pages/Login';
import { Dashboard } from '@/pages/Dashboard';
import { EnvironmentDetails } from '@/pages/EnvironmentDetails';
import { CreateEnvironment } from '@/pages/CreateEnvironment';
import { EditEnvironment } from '@/pages/EditEnvironment';
import { Logs } from '@/pages/Logs';
import { Users } from '@/pages/Users';
import { NotFound } from '@/pages/NotFound';

const PrivateRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = localStorage.getItem('token');
  return token ? <>{children}</> : <Navigate to="/login" />;
};

export const AppRoutes: React.FC = () => {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <PrivateRoute>
            <MainLayout />
          </PrivateRoute>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="environments/create" element={<CreateEnvironment />} />
        <Route path="environments/:id" element={<EnvironmentDetails />} />
        <Route path="environments/:id/edit" element={<EditEnvironment />} />
        <Route path="logs" element={<Logs />} />
        <Route path="users" element={<Users />} />
      </Route>
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
};
```

## Performance Optimizations

1. **Code Splitting**: Routes are automatically code-split by Vite
2. **React.memo**: Used for expensive components like EnvironmentCard
3. **Redux Toolkit**: Built-in performance optimizations with Immer
4. **Virtual Scrolling**: For large lists (logs, environments)
5. **Debounced Search**: For filtering operations
6. **Optimistic Updates**: For better UX on actions

## Build and Deployment

The application is containerized with multi-stage Docker builds:

```dockerfile
# frontend/Dockerfile
FROM node:18-alpine as builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
