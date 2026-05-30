import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
import { EditEnvironment } from '../EditEnvironment';
import { environmentApi } from '@/api/environments';
import { HealthStatus } from '@/types/environment';

const mockNavigate = vi.fn();
const mockEnqueueSnackbar = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ id: 'env-1' }),
  };
});

vi.mock('notistack', () => ({
  useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar, closeSnackbar: vi.fn() }),
  SnackbarProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

vi.mock('@/api/environments');
const mockedEnvApi = vi.mocked(environmentApi);

// Mock EnvironmentForm so we can trigger onSubmit directly
vi.mock('@/components/environments/EnvironmentForm', () => ({
  EnvironmentForm: ({
    onSubmit,
    error,
  }: {
    onSubmit: (data: any, password?: string, privateKey?: string) => Promise<void>;
    error?: string | null;
  }) => {
    const baseData = {
      name: 'test-env',
      description: '',
      environmentURL: 'https://test.example.com',
      target: { host: 'localhost', port: 22 },
      healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
      commands: { type: 'ssh', restart: { enabled: false } },
      upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
    };
    return (
      <div>
        {error && <div data-testid="form-error">{error}</div>}
        <button
          data-testid="submit-password"
          onClick={() =>
            onSubmit({ ...baseData, credentials: { type: 'password', username: 'user' } }, 'secret123', undefined)
          }
        >
          Submit Password
        </button>
        <button
          data-testid="submit-key"
          onClick={() =>
            onSubmit({ ...baseData, credentials: { type: 'key', username: 'user' } }, undefined, '---key---')
          }
        >
          Submit Key
        </button>
        <button
          data-testid="submit-no-creds"
          onClick={() => onSubmit({ ...baseData, credentials: { type: 'password', username: 'user' } })}
        >
          Submit No Creds
        </button>
      </div>
    );
  },
}));

const mockEnvironment = {
  id: 'env-1',
  name: 'Test Env',
  description: 'desc',
  environmentURL: 'https://test.example.com',
  target: { host: 'localhost', port: 22 },
  credentials: { type: 'password' as const, username: 'user' },
  healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode' as const, value: 200 } },
  status: { health: HealthStatus.Unknown, lastCheck: new Date().toISOString(), message: '', responseTime: 0 },
  systemInfo: { osVersion: '', appVersion: '', lastUpdated: new Date().toISOString() },
  timestamps: { createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  commands: { type: 'ssh' as const, restart: { enabled: false } },
  upgradeConfig: { enabled: false, type: 'ssh' as const, upgradeCommand: {} },
};

describe('EditEnvironment — mutation paths', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
  });

  it('calls update API with password credentials and navigates on success', async () => {
    mockedEnvApi.update = vi.fn().mockResolvedValue(mockEnvironment);
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByTestId('submit-password')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('submit-password'));

    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalledWith(
        'env-1',
        expect.objectContaining({
          credentials: { type: 'password', username: 'user' },
          metadata: { password: 'secret123' },
        })
      );
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
      expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Environment updated successfully', { variant: 'success' });
    });
  });

  it('calls update API with privateKey credentials on success', async () => {
    mockedEnvApi.update = vi.fn().mockResolvedValue(mockEnvironment);
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByTestId('submit-key')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('submit-key'));

    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalledWith(
        'env-1',
        expect.objectContaining({
          metadata: { privateKey: '---key---' },
        })
      );
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });
  });

  it('shows error message when update fails', async () => {
    const apiError = { response: { data: { message: 'Update failed' } } };
    mockedEnvApi.update = vi.fn().mockRejectedValue(apiError);
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByTestId('submit-no-creds')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('submit-no-creds'));

    await waitFor(() => {
      expect(screen.getByTestId('form-error')).toHaveTextContent('Update failed');
    });
  });

  it('shows fallback error when update fails without response data', async () => {
    mockedEnvApi.update = vi.fn().mockRejectedValue(new Error('Network error'));
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByTestId('submit-no-creds')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('submit-no-creds'));

    await waitFor(() => {
      expect(screen.getByTestId('form-error')).toHaveTextContent('Failed to update environment');
    });
  });

  it('does not include metadata when no password or key is provided', async () => {
    mockedEnvApi.update = vi.fn().mockResolvedValue(mockEnvironment);
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByTestId('submit-no-creds')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('submit-no-creds'));

    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalledWith(
        'env-1',
        expect.not.objectContaining({ metadata: expect.anything() })
      );
    });
  });
});
