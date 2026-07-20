import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
import { NotFound } from '../NotFound';
import { CreateEnvironment } from '../CreateEnvironment';
import { EditEnvironment } from '../EditEnvironment';
import { environmentApi } from '@/api/environments';
import { HealthStatus } from '@/types/environment';

// Mock notistack used in CreateEnvironment/EditEnvironment
const mockEnqueueSnackbar = vi.fn();
vi.mock('notistack', () => ({
  useSnackbar: () => ({
    enqueueSnackbar: mockEnqueueSnackbar,
    closeSnackbar: vi.fn(),
  }),
  SnackbarProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

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

// Mock EnvironmentForm to simplify testing page wrapper mutations & logic
vi.mock('@/components/environments/EnvironmentForm', () => ({
  EnvironmentForm: ({ mode, onSubmit, isLoading, error }: any) => {
    const handleSub = async (data: any, pwd?: string, key?: string) => {
      try {
        await onSubmit(data, pwd, key);
      } catch (err) {
        // expected rejected promise, caught here to prevent unhandled rejection
      }
    };
    return (
      <div data-testid="mock-env-form">
        <span>Mode: {mode}</span>
        <span>IsLoading: {isLoading ? 'yes' : 'no'}</span>
        {error && <span data-testid="form-error">{error}</span>}
        <button
          data-testid="submit-btn-password"
          onClick={() => handleSub({
            name: 'Updated Env Password',
            description: 'Updated Desc',
            environmentURL: 'https://updated.example.com',
            target: { host: 'localhost', port: 22 },
            credentials: { type: 'password', username: 'user' },
            healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
            commands: { type: 'ssh', restart: { enabled: false } },
            upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
          }, 'new-password')}
        >
          Submit Form Password
        </button>
        <button
          data-testid="submit-btn-key"
          onClick={() => handleSub({
            name: 'Updated Env Key',
            description: 'Updated Desc',
            environmentURL: 'https://updated.example.com',
            target: { host: 'localhost', port: 22 },
            credentials: { type: 'key', username: 'user' },
            healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
            commands: { type: 'ssh', restart: { enabled: false } },
            upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
          }, undefined, 'new-private-key')}
        >
          Submit Form Key
        </button>
      </div>
    );
  }
}));

const mockEnvironment = {
  id: 'env-1',
  name: 'Test Env',
  description: 'desc',
  target: { host: 'localhost', port: 22 },
  credentials: { type: 'password', username: 'user' },
  healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
  status: { health: HealthStatus.Unknown, lastCheck: new Date().toISOString(), message: '', responseTime: 0 },
  systemInfo: { osVersion: '', appVersion: '', lastUpdated: new Date().toISOString() },
  timestamps: { createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  commands: { type: 'ssh', restart: { enabled: false } },
  upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
};

describe('NotFound page', () => {
  it('renders 404 page', () => {
    render(<NotFound />);
    expect(screen.getByText('404')).toBeInTheDocument();
  });

  it('shows page not found text', () => {
    render(<NotFound />);
    expect(screen.getByText(/page not found/i)).toBeInTheDocument();
  });

  it('has go to dashboard button', () => {
    render(<NotFound />);
    expect(screen.getByRole('button', { name: /go to dashboard/i })).toBeInTheDocument();
  });
});

describe('CreateEnvironment page', () => {
  it('renders create environment heading', () => {
    render(<CreateEnvironment />);
    expect(screen.getByText(/create.*new.*environment/i)).toBeInTheDocument();
  });

  it('renders the environment form', () => {
    render(<CreateEnvironment />);
    expect(screen.getByTestId('mock-env-form')).toBeInTheDocument();
  });

  it('calls environmentApi.create on form submission with password', async () => {
    mockedEnvApi.create = vi.fn().mockResolvedValue({ id: 'new-env' });
    render(<CreateEnvironment />);

    fireEvent.click(screen.getByTestId('submit-btn-password'));

    await waitFor(() => {
      expect(mockedEnvApi.create).toHaveBeenCalledWith({
        name: 'Updated Env Password',
        description: 'Updated Desc',
        environmentURL: 'https://updated.example.com',
        target: { host: 'localhost', port: 22 },
        credentials: { type: 'password', username: 'user' },
        healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
        commands: { type: 'ssh', restart: { enabled: false } },
        upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
        metadata: {
          password: 'new-password'
        }
      });
      expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Environment created successfully', { variant: 'success' });
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });
  });

  it('calls environmentApi.create on form submission with key', async () => {
    mockedEnvApi.create = vi.fn().mockResolvedValue({ id: 'new-env' });
    render(<CreateEnvironment />);

    fireEvent.click(screen.getByTestId('submit-btn-key'));

    await waitFor(() => {
      expect(mockedEnvApi.create).toHaveBeenCalledWith({
        name: 'Updated Env Key',
        description: 'Updated Desc',
        environmentURL: 'https://updated.example.com',
        target: { host: 'localhost', port: 22 },
        credentials: { type: 'key', username: 'user' },
        healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
        commands: { type: 'ssh', restart: { enabled: false } },
        upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
        metadata: {
          privateKey: 'new-private-key'
        }
      });
    });
  });

  it('shows error on failed creation', async () => {
    const errorMsg = 'Invalid hostname';
    mockedEnvApi.create = vi.fn().mockRejectedValue({
      response: { data: { message: errorMsg } }
    });
    render(<CreateEnvironment />);

    fireEvent.click(screen.getByTestId('submit-btn-password'));

    await waitFor(() => {
      expect(screen.getByTestId('form-error')).toHaveTextContent(errorMsg);
    });
  });

  it('shows generic error on failed creation without specific message', async () => {
    mockedEnvApi.create = vi.fn().mockRejectedValue({});
    render(<CreateEnvironment />);

    fireEvent.click(screen.getByTestId('submit-btn-password'));

    await waitFor(() => {
      expect(screen.getByTestId('form-error')).toHaveTextContent('Failed to create environment');
    });
  });
});

describe('EditEnvironment page', () => {
  it('shows loading spinner while fetching', () => {
    mockedEnvApi.get = vi.fn(() => new Promise(() => {}));
    render(<EditEnvironment />);
    expect(document.querySelector('.MuiCircularProgress-root')).not.toBeNull();
  });

  it('renders edit form when loaded', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    render(<EditEnvironment />);
    await waitFor(() => {
      expect(screen.getByText(/edit environment/i)).toBeInTheDocument();
    });
  });

  it('shows environment not found when environment missing', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(null);
    render(<EditEnvironment />);
    await waitFor(() => {
      expect(screen.getByText(/environment not found/i)).toBeInTheDocument();
    });
  });

  it('shows environment name in heading when loaded', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    render(<EditEnvironment />);
    await waitFor(() => {
      expect(screen.getByText(/test env/i)).toBeInTheDocument();
    });
  });

  it('submits updated data successfully with password', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockResolvedValue({ id: 'env-1' });
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByTestId('submit-btn-password')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('submit-btn-password'));

    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalledWith('env-1', {
        name: 'Updated Env Password',
        description: 'Updated Desc',
        environmentURL: 'https://updated.example.com',
        target: { host: 'localhost', port: 22 },
        credentials: { type: 'password', username: 'user' },
        healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
        commands: { type: 'ssh', restart: { enabled: false } },
        upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
        metadata: {
          password: 'new-password'
        }
      });
      expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Environment updated successfully', { variant: 'success' });
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });
  });

  it('submits updated data successfully with key', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockResolvedValue({ id: 'env-1' });
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByTestId('submit-btn-key')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('submit-btn-key'));

    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalledWith('env-1', {
        name: 'Updated Env Key',
        description: 'Updated Desc',
        environmentURL: 'https://updated.example.com',
        target: { host: 'localhost', port: 22 },
        credentials: { type: 'key', username: 'user' },
        healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
        commands: { type: 'ssh', restart: { enabled: false } },
        upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
        metadata: {
          privateKey: 'new-private-key'
        }
      });
    });
  });

  it('shows error on failed update', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    const errorMsg = 'Update failed database constraint';
    mockedEnvApi.update = vi.fn().mockRejectedValue({
      response: { data: { message: errorMsg } }
    });
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByTestId('submit-btn-password')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('submit-btn-password'));

    await waitFor(() => {
      expect(screen.getByTestId('form-error')).toHaveTextContent(errorMsg);
    });
  });
});
