import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@/test/test-utils';
import { AppRoutes } from '../index';

// Mock all page components to avoid full render complexity
vi.mock('@/pages/Dashboard', () => ({
  Dashboard: () => <div data-testid="dashboard-page">Dashboard</div>,
}));
vi.mock('@/pages/Login', () => ({
  Login: () => <div data-testid="login-page">Login</div>,
}));
vi.mock('@/pages/EnvironmentDetails', () => ({
  EnvironmentDetails: () => <div data-testid="env-details-page">Env Details</div>,
}));
vi.mock('@/pages/CreateEnvironment', () => ({
  CreateEnvironment: () => <div data-testid="create-env-page">Create Env</div>,
}));
vi.mock('@/pages/EditEnvironment', () => ({
  EditEnvironment: () => <div data-testid="edit-env-page">Edit Env</div>,
}));
vi.mock('@/pages/Logs', () => ({
  Logs: () => <div data-testid="logs-page">Logs</div>,
}));
vi.mock('@/pages/NotFound', () => ({
  NotFound: () => <div data-testid="not-found-page">Not Found</div>,
}));
vi.mock('@/pages/Users', () => ({
  default: () => <div data-testid="users-page">Users</div>,
}));
vi.mock('@/layouts/MainLayout', () => ({
  MainLayout: () => <div data-testid="main-layout">Main Layout</div>,
}));

describe('AppRoutes', () => {
  beforeEach(() => {
    localStorage.removeItem('authToken');
    vi.clearAllMocks();
  });

  it('renders without crashing', () => {
    const { container } = render(<AppRoutes />);
    expect(container).toBeTruthy();
  });

  it('renders login page at /login route when not authenticated', async () => {
    // The router in test-utils uses BrowserRouter, window.location is /
    // No authToken in localStorage — ProtectedRoute redirects to /login
    render(<AppRoutes />);
    await waitFor(() => {
      // Since BrowserRouter starts at "/" and no auth token, should redirect to /login
      expect(screen.getByTestId('login-page')).toBeInTheDocument();
    });
  });

  it('renders login page by default (unauthenticated)', async () => {
    render(<AppRoutes />);
    await waitFor(() => {
      // Without auth token, ProtectedRoute redirects to /login
      expect(screen.getByTestId('login-page')).toBeInTheDocument();
    });
  });
});
