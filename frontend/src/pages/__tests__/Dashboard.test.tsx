import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
import { Dashboard } from '../Dashboard';
import { environmentApi } from '@/api/environments';
import { HealthStatus } from '@/types/environment';

vi.mock('@/api/environments');
const mockedEnvApi = vi.mocked(environmentApi);

const mockEnvironment = {
  id: 'env-1',
  name: 'Production',
  description: 'Prod env',
  target: { host: 'prod.example.com', port: 22 },
  credentials: { type: 'password', username: 'user' },
  healthCheck: { enabled: true, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
  status: { health: HealthStatus.Healthy, lastCheck: new Date().toISOString(), message: 'ok', responseTime: 50 },
  systemInfo: { osVersion: 'Ubuntu 22.04', appVersion: '2.0.0', lastUpdated: new Date().toISOString() },
  timestamps: { createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  commands: { type: 'ssh', restart: { enabled: true, command: 'restart' } },
  upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
};

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

describe('Dashboard page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows loading skeletons while fetching', () => {
    mockedEnvApi.list = vi.fn(() => new Promise(() => {})); // Never resolves
    render(<Dashboard />);
    // Skeletons are rendered during loading
    const skeletons = document.querySelectorAll('.MuiSkeleton-root');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('renders environments when loaded', async () => {
    mockedEnvApi.list = vi.fn().mockResolvedValue({
      environments: [mockEnvironment],
      pagination: { page: 1, limit: 10, total: 1 },
    });

    render(<Dashboard />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });
  });

  it('shows empty state when no environments', async () => {
    mockedEnvApi.list = vi.fn().mockResolvedValue({
      environments: [],
      pagination: { page: 1, limit: 10, total: 0 },
    });

    render(<Dashboard />);
    await waitFor(() => {
      expect(screen.getByText(/no environments configured/i)).toBeInTheDocument();
    });
  });

  it('shows error when fetch fails', async () => {
    mockedEnvApi.list = vi.fn().mockRejectedValue(new Error('Network error'));
    render(<Dashboard />);
    await waitFor(() => {
      expect(screen.getByText(/failed to load environments/i)).toBeInTheDocument();
    });
  });

  it('shows stat cards with counts', async () => {
    const unhealthyEnv = { ...mockEnvironment, id: 'env-2', status: { ...mockEnvironment.status, health: HealthStatus.Unhealthy } };
    const unknownEnv = { ...mockEnvironment, id: 'env-3', status: { ...mockEnvironment.status, health: HealthStatus.Unknown } };
    mockedEnvApi.list = vi.fn().mockResolvedValue({
      environments: [mockEnvironment, unhealthyEnv, unknownEnv],
      pagination: { page: 1, limit: 10, total: 3 },
    });

    render(<Dashboard />);
    await waitFor(() => {
      expect(screen.getByText('Total Environments')).toBeInTheDocument();
      expect(screen.getByText('Healthy')).toBeInTheDocument();
      expect(screen.getByText('Unhealthy')).toBeInTheDocument();
      expect(screen.getByText('Unknown')).toBeInTheDocument();
    });
  });

  it('navigates to create environment on "New Environment" click', async () => {
    mockedEnvApi.list = vi.fn().mockResolvedValue({
      environments: [],
      pagination: { page: 1, limit: 10, total: 0 },
    });

    render(<Dashboard />);
    await waitFor(() => {
      expect(screen.getByText(/no environments configured/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /new environment/i }));
    expect(mockNavigate).toHaveBeenCalledWith('/environments/create');
  });

  it('navigates to create on empty state button click', async () => {
    mockedEnvApi.list = vi.fn().mockResolvedValue({
      environments: [],
      pagination: { page: 1, limit: 10, total: 0 },
    });

    render(<Dashboard />);
    await waitFor(() => {
      expect(screen.getByText(/create your first environment/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /create your first environment/i }));
    expect(mockNavigate).toHaveBeenCalledWith('/environments/create');
  });
});
