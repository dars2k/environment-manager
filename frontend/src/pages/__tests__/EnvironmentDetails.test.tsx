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
const { mockDeleteEnvironment } = vi.hoisted(() => ({
  mockDeleteEnvironment: vi.fn(),
}));
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

// Environment exercising fallback / "not configured" branches: missing system
// info, a zero response time, HTTP-type restart & upgrade commands with no
// url/method configured.
const mockEnvironmentEdgeCases = {
  ...mockEnvironment,
  systemInfo: { osVersion: 'Ubuntu 22.04', appVersion: '', lastUpdated: '' },
  status: { ...mockEnvironment.status, responseTime: 0 },
  commands: { type: 'http', restart: { enabled: true, url: '', method: undefined } },
  upgradeConfig: {
    enabled: true,
    type: 'http',
    versionListURL: '',
    jsonPathResponse: '',
    upgradeCommand: { url: '', method: undefined },
  },
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

  it('calls deleteEnvironment and navigates to dashboard when the delete icon is clicked', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    const user = userEvent.setup();
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    const deleteButton = screen.getByRole('button', { name: /delete environment/i });
    await user.click(deleteButton);

    expect(mockDeleteEnvironment).toHaveBeenCalledWith('env-1');
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
  });

  it('shows "Not available" fallback text for missing system info fields', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironmentEdgeCases);
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    // System Info tab is shown by default (tabValue starts at 0). There are
    // three occurrences: the missing appVersion, the missing lastUpdated,
    // and the static "Not available" example inside the info Alert copy.
    const notAvailable = screen.getAllByText('Not available');
    expect(notAvailable).toHaveLength(3);
    expect(screen.getByText('Ubuntu 22.04')).toBeInTheDocument();
  });

  it('shows N/A for the response time when it is zero', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironmentEdgeCases);
    const user = userEvent.setup();
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    await user.click(screen.getByRole('tab', { name: /health check/i }));

    expect(screen.getByText('N/A')).toBeInTheDocument();
  });

  it('renders HTTP restart command fallback fields and the missing-command warning', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironmentEdgeCases);
    const user = userEvent.setup();
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    await user.click(screen.getByRole('tab', { name: /commands/i }));

    // HTTP endpoint not configured
    expect(screen.getAllByText('Not configured').length).toBeGreaterThan(0);
    // Method falls back to POST
    expect(screen.getByText('POST')).toBeInTheDocument();
    // Warning shown because neither command nor url is set
    expect(
      screen.getByText(/no restart command configured/i)
    ).toBeInTheDocument();
  });

  it('renders HTTP upgrade command fallback fields when upgrade config values are unconfigured', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironmentEdgeCases);
    const user = userEvent.setup();
    render(<EnvironmentDetails />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    await user.click(screen.getByRole('tab', { name: /upgrade config/i }));

    // Version List URL, JSONPath Response, and HTTP Endpoint all fall back
    expect(screen.getAllByText('Not configured').length).toBe(3);
    // Upgrade command method falls back to POST
    expect(screen.getByText('POST')).toBeInTheDocument();
  });
});
