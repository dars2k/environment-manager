# Frontend Architecture Design

## Project Structure

```
frontend/
├── public/
├── src/
│   ├── api/                   # API service layer
│   │   ├── environments.ts    # Environment API calls
│   │   └── users.ts           # User API calls
│   ├── components/            # Reusable UI components
│   │   ├── common/
│   │   │   └── ConfirmDialog.tsx
│   │   ├── environments/
│   │   │   ├── CreateEnvironmentDialog.tsx
│   │   │   ├── EditEnvironmentDialog.tsx
│   │   │   ├── EnvironmentCard.tsx
│   │   │   ├── EnvironmentForm.tsx
│   │   │   ├── EnvironmentLogs.tsx
│   │   │   ├── HttpRequestConfig.tsx
│   │   │   ├── UpgradeEnvironmentDialog.tsx
│   │   │   └── index.ts
│   │   └── layout/
│   ├── contexts/
│   │   └── WebSocketContext.tsx
│   ├── hooks/
│   │   └── useEnvironmentActions.ts
│   ├── layouts/
│   │   └── MainLayout.tsx
│   ├── pages/
│   │   ├── Dashboard.tsx
│   │   ├── EnvironmentDetails.tsx
│   │   ├── Login.tsx
│   │   ├── Logs.tsx
│   │   ├── NotFound.tsx
│   │   └── Users.tsx
│   ├── routes/
│   │   └── index.tsx
│   ├── store/
│   │   ├── index.ts
│   │   └── slices/
│   │       ├── environmentSlice.ts
│   │       ├── logsSlice.ts
│   │       ├── notificationSlice.ts
│   │       └── uiSlice.ts
│   ├── theme/
│   │   └── index.ts
│   ├── App.tsx
│   └── main.tsx
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

## Core Technologies

| Technology | Version | Purpose |
|-----------|---------|---------|
| React | 18 | UI framework |
| TypeScript | 5 | Type safety |
| Redux Toolkit | latest | Global state management |
| Material-UI | v5 | UI component library |
| React Router | v6 | Client-side routing |
| Vite | latest | Build tool and dev server |
| Axios | latest | HTTP client |
| Native WebSocket | — | Real-time updates |

## State Management

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
  initialState: { environments: [], selectedEnvironment: null, loading: false, error: null } as EnvironmentState,
  reducers: {
    updateEnvironmentStatus: (state, action: PayloadAction<{ id: string; status: Environment['status'] }>) => {
      const env = state.environments.find(e => e.id === action.payload.id);
      if (env) env.status = action.payload.status;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchEnvironments.pending, (state) => { state.loading = true; state.error = null; })
      .addCase(fetchEnvironments.fulfilled, (state, action) => { state.loading = false; state.environments = action.payload; })
      .addCase(fetchEnvironments.rejected, (state, action) => { state.loading = false; state.error = action.error.message || 'Failed to fetch environments'; });
  },
});
```

## API Service Layer

All API calls go through a shared Axios instance that injects the JWT token. The base URL is `/api/v1` — in production, nginx proxies this to the backend; in development, Vite's proxy handles it.

```typescript
// src/api/environments.ts
import axios from 'axios';

const apiClient = axios.create({ baseURL: '/api/v1' });

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export const environmentsApi = {
  getAll: () => apiClient.get<Environment[]>('/environments'),
  getById: (id: string) => apiClient.get<Environment>(`/environments/${id}`),
  create: (data: Partial<Environment>) => apiClient.post<Environment>('/environments', data),
  update: (id: string, data: Partial<Environment>) => apiClient.put<Environment>(`/environments/${id}`, data),
  delete: (id: string) => apiClient.delete(`/environments/${id}`),
  restart: (id: string, data?: { force?: boolean }) => apiClient.post(`/environments/${id}/restart`, data),
  checkHealth: (id: string) => apiClient.post(`/environments/${id}/check-health`),
  getVersions: (id: string) => apiClient.get(`/environments/${id}/versions`),
  upgrade: (id: string, data: { version: string }) => apiClient.post(`/environments/${id}/upgrade`, data),
  getLogs: (id: string, params?: Record<string, unknown>) => apiClient.get(`/environments/${id}/logs`, { params }),
};
```

## WebSocket Integration

```typescript
// src/contexts/WebSocketContext.tsx
export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const dispatch = useDispatch();
  const ws = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = React.useState(false);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) return;

    // Path is proxied by nginx to the backend WebSocket
    ws.current = new WebSocket(`${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`);

    ws.current.onopen = () => setIsConnected(true);
    ws.current.onclose = () => setIsConnected(false);

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
          dispatch(addNotification({ type: 'info', message: `Operation ${message.payload.status}` }));
          break;
      }
    };

    return () => ws.current?.close();
  }, [dispatch]);

  // ...
};
```

## Routing and RBAC

Routes are protected by role. The `MainLayout` reads the current user's role from the Redux store and conditionally renders navigation items and action buttons.

```tsx
// src/routes/index.tsx
const PrivateRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = localStorage.getItem('token');
  return token ? <>{children}</> : <Navigate to="/login" />;
};

export const AppRoutes: React.FC = () => (
  <Routes>
    <Route path="/login" element={<Login />} />
    <Route path="/" element={<PrivateRoute><MainLayout /></PrivateRoute>}>
      <Route index element={<Dashboard />} />
      <Route path="environments/:id" element={<EnvironmentDetails />} />
      <Route path="logs" element={<Logs />} />
      <Route path="users" element={<Users />} />   {/* admin only — hidden in nav for non-admins */}
    </Route>
    <Route path="*" element={<NotFound />} />
  </Routes>
);
```

### Role-gated UI elements

- **Admin-only**: "New Environment" button, Edit/Delete env icons, Users nav item, Create User / Edit User dialogs
- **All roles**: Dashboard, environment detail view, Restart / Upgrade buttons, logs

## Custom Hooks

```typescript
// src/hooks/useEnvironmentActions.ts
export const useEnvironmentActions = (environmentId: string) => {
  const dispatch = useDispatch<AppDispatch>();
  const [loading, setLoading] = useState(false);

  const restart = async (force = false) => {
    setLoading(true);
    try {
      await environmentsApi.restart(environmentId, { force });
      dispatch(addNotification({ type: 'success', message: 'Environment restart initiated' }));
    } catch {
      dispatch(addNotification({ type: 'error', message: 'Failed to restart environment' }));
    } finally {
      setLoading(false);
    }
  };

  const upgrade = async (version: string) => {
    setLoading(true);
    try {
      await environmentsApi.upgrade(environmentId, { version });
      dispatch(addNotification({ type: 'success', message: 'Upgrade initiated' }));
    } catch {
      dispatch(addNotification({ type: 'error', message: 'Failed to upgrade environment' }));
    } finally {
      setLoading(false);
    }
  };

  return { restart, upgrade, loading };
};
```

## Theme

```typescript
// src/theme/index.ts
export const theme = createTheme({
  palette: {
    mode: 'dark',
    primary:    { main: '#3b82f6' },
    secondary:  { main: '#8b5cf6' },
    success:    { main: '#10b981' },
    error:      { main: '#ef4444' },
    warning:    { main: '#f59e0b' },
    background: { default: '#0f172a', paper: '#1e293b' },
  },
  typography: {
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
  },
  components: {
    MuiButton: { styleOverrides: { root: { textTransform: 'none' } } },
    MuiCard:   { styleOverrides: { root: { borderRadius: 8 } } },
  },
});
```

## Build and Deployment

Multi-stage Docker build:

```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:stable-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

Nginx proxies `/api/*` and `/ws` to the backend, and serves the React SPA for all other paths.

## Performance

- **Code splitting** — Vite automatically splits routes into separate chunks
- **React.memo** — applied to frequently re-rendered cards (e.g., `EnvironmentCard`)
- **Redux Toolkit** — uses Immer for efficient immutable state updates
- **Debounced search** — 300ms debounce on log/environment filter inputs
- **Optimistic updates** — UI state updated immediately on user actions before server confirmation
