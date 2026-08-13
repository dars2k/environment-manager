import { describe, it, expect, vi, beforeEach } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '@/test/test-utils';
import { EnvironmentDetails } from '../EnvironmentDetails';
import { environmentApi } from '@/api/environments';
import { HealthStatus } from '@/types/environment';

vi.mock('@/api/environments');
const mockedEnvApi = vi.mocked(environmentApi);

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ id: 'env-1' }),
  };
});

// Mock useEnvironmentActions hook
const mockDeleteEnvironment = vi.fn();
vi.mock('@/hooks/useEnvironmentActions', () => ({
  useEnvironmentActions: () => ({
    deleteEnvironment: mockDeleteEnvironment,
    restartEnvironment: vi.fn(),
    upgradeEnvironment: vi.fn(),
    checkHealth: vi.fn(),
  }),
}));

const mockEnvironment = {
  id: 'env-1',
  name: 'Production',
  description: 'Production environment',
  environmentURL: 'https://prod.example.com',
  target: { host: 'prod.example.com', port: 22 },
  credentials: { type: 'password', username: 'deploy' },
  healthCheck: { enabled: true, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
  status: { health: HealthStatus.Healthy, lastCheck: new Date().toISOString(), message: 'OK', responseTime: 100 },
  systemInfo: { osVersion: 'Ubuntu 22.04', appVersion: '3.1.0', lastUpdated: new Date().toISOString() },
  timestamps: { createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  commands: { type: 'ssh', restart: { enabled: true, command: 'restart' } },
  upgradeConfig: { enabled: true, type: 'ssh', versionListURL: 'https://api.example.com/versions', jsonPathResponse: '$.versions', upgradeCommand: {} },
};

describe('EnvironmentDetails page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows loading spinner while fetching', () => {
    mockedEnvApi.get = vi.fn(() => new Promise(() => {}));
    render(<EnvironmentDetails />);
    expect(document.querySelector('.MuiCircularProgress-root')).not.toBeNull();
  });

  it('shows error when environment not found', async () => {
    mockedEnvApi.get = vi.fn().mockRejectedValue(new Error('Not found'));
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText(/failed to load environment/i)).toBeInTheDocument();
    });
  });

  it('renders environment name when loaded', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });
  });

  it('shows health status chip', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText(/healthy/i)).toBeInTheDocument();
    });
  });

  it('has back to dashboard button on error', async () => {
    mockedEnvApi.get = vi.fn().mockRejectedValue(new Error('Network error'));
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /back to dashboard/i })).toBeInTheDocument();
    });
  });

  it('renders environment URL when available', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('https://prod.example.com')).toBeInTheDocument();
    });
  });

  it('shows unhealthy status', async () => {
    const unhealthyEnv = { ...mockEnvironment, status: { ...mockEnvironment.status, health: HealthStatus.Unhealthy } };
    mockedEnvApi.get = vi.fn().mockResolvedValue(unhealthyEnv);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText(/unhealthy/i)).toBeInTheDocument();
    });
  });

  it('shows unknown status', async () => {
    const unknownEnv = { ...mockEnvironment, status: { ...mockEnvironment.status, health: HealthStatus.Unknown } };
    mockedEnvApi.get = vi.fn().mockResolvedValue(unknownEnv);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText(/unknown/i)).toBeInTheDocument();
    });
  });

  it('deletes the environment and navigates to the dashboard when delete button clicked', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: /delete environment/i }));

    expect(mockDeleteEnvironment).toHaveBeenCalledWith('env-1');
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
  });

  it('shows "Not available" for missing app version and last system update', async () => {
    const envMissingSystemInfo = {
      ...mockEnvironment,
      systemInfo: { osVersion: 'Ubuntu 22.04', appVersion: '', lastUpdated: '' },
    };
    mockedEnvApi.get = vi.fn().mockResolvedValue(envMissingSystemInfo);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    // System Info tab is the default tab (index 0), so no click needed.
    const notAvailable = screen.getAllByText('Not available');
    // One for Application Version, one for Last System Update, plus the
    // static hint text inside the info Alert ("...show Not available until...").
    expect(notAvailable).toHaveLength(3);
  });

  it('shows "N/A" for zero response time on the Health Check tab', async () => {
    const envZeroResponseTime = {
      ...mockEnvironment,
      status: { ...mockEnvironment.status, responseTime: 0 },
    };
    mockedEnvApi.get = vi.fn().mockResolvedValue(envZeroResponseTime);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    const user = userEvent.setup();
    await user.click(screen.getByRole('tab', { name: /health check/i }));

    expect(await screen.findByText('N/A')).toBeInTheDocument();
  });

  it('renders HTTP restart command fields and the missing-command warning on the Commands tab', async () => {
    const envHttpCommands = {
      ...mockEnvironment,
      commands: {
        type: 'http' as const,
        restart: { enabled: true, url: '', method: '' },
      },
    };
    mockedEnvApi.get = vi.fn().mockResolvedValue(envHttpCommands);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    const user = userEvent.setup();
    await user.click(screen.getByRole('tab', { name: /commands/i }));

    expect(await screen.findByText('HTTP Endpoint')).toBeInTheDocument();
    // Falls back to "Not configured" since url is empty.
    expect(screen.getAllByText('Not configured').length).toBeGreaterThan(0);
    // Falls back to POST since method is empty.
    expect(screen.getByText('POST')).toBeInTheDocument();
    // Neither command nor url is configured -> warning alert shown.
    expect(screen.getByText(/no restart command configured/i)).toBeInTheDocument();
  });

  it('renders HTTP upgrade command fields and empty version list/JSONPath fallbacks on the Upgrade Config tab', async () => {
    const envHttpUpgrade = {
      ...mockEnvironment,
      upgradeConfig: {
        enabled: true,
        type: 'http' as const,
        versionListURL: '',
        jsonPathResponse: '',
        upgradeCommand: { url: '', method: '' },
      },
    };
    mockedEnvApi.get = vi.fn().mockResolvedValue(envHttpUpgrade);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    const user = userEvent.setup();
    await user.click(screen.getByRole('tab', { name: /upgrade config/i }));

    expect(await screen.findByText('Version List URL')).toBeInTheDocument();
    expect(screen.getByText('JSONPath Response')).toBeInTheDocument();
    expect(screen.getByText('HTTP Endpoint')).toBeInTheDocument();
    // versionListURL, jsonPathResponse and upgradeCommand.url are all empty.
    expect(screen.getAllByText('Not configured').length).toBe(3);
    // upgradeCommand.method is empty -> falls back to POST.
    expect(screen.getByText('POST')).toBeInTheDocument();
  });
});
