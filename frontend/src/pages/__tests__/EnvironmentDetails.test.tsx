import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
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

vi.mock('@/hooks/useEnvironmentActions', () => ({
  useEnvironmentActions: () => ({
    deleteEnvironment: vi.fn(),
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

  it('renders HTTP upgrade config when enabled', async () => {
    const httpUpgradeEnv = {
      ...mockEnvironment,
      upgradeConfig: {
        enabled: true,
        type: 'http',
        versionListURL: 'https://versions.example.com',
        jsonPathResponse: '$.latest',
        upgradeCommand: {
          url: 'https://upgrade.example.com',
          method: 'PUT'
        }
      }
    };
    mockedEnvApi.get = vi.fn().mockResolvedValue(httpUpgradeEnv);
    render(<EnvironmentDetails />);

    await waitFor(() => expect(screen.getByText('Production')).toBeInTheDocument());
    fireEvent.click(screen.getByText('Upgrade Config'));

    await waitFor(() => {
      expect(screen.getByText('https://upgrade.example.com')).toBeInTheDocument();
      expect(screen.getByText('PUT')).toBeInTheDocument();
      expect(screen.getByText('https://versions.example.com')).toBeInTheDocument();
      expect(screen.getByText('$.latest')).toBeInTheDocument();
    });
  });
});
