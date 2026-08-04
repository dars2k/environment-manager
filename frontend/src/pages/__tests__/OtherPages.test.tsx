import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@/test/test-utils';
import { NotFound } from '../NotFound';
import { CreateEnvironment } from '../CreateEnvironment';
import { EditEnvironment } from '../EditEnvironment';
import { environmentApi } from '@/api/environments';
import { HealthStatus } from '@/types/environment';

// Mock notistack used in CreateEnvironment/EditEnvironment
vi.mock('notistack', () => ({
  useSnackbar: () => ({
    enqueueSnackbar: vi.fn(),
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

// Mock EnvironmentForm to allow direct unit testing of submit logic and states
vi.mock('@/components/environments/EnvironmentForm', () => ({
  EnvironmentForm: ({ mode, initialData, onSubmit, isLoading, error }: any) => {
    const [name, setName] = React.useState(initialData?.name || '');
    const [url, setUrl] = React.useState(initialData?.environmentURL || '');
    const [credType, setCredType] = React.useState(initialData?.credentials?.type || 'password');

    return (
      <div data-testid="mock-environment-form">
        {error && <div data-testid="form-error">{error}</div>}
        {isLoading && <div data-testid="form-loading">Loading...</div>}
        <label>
          Environment Name *
          <input value={name} onChange={(e) => setName(e.target.value)} />
        </label>
        <label>
          Environment URL
          <input value={url} onChange={(e) => setUrl(e.target.value)} />
        </label>
        <select data-testid="cred-type-select" value={credType} onChange={(e) => setCredType(e.target.value)}>
          <option value="password">Password</option>
          <option value="key">Key</option>
        </select>
        <button
          data-testid="submit-form-btn"
          onClick={async () => {
            const testData = {
              name,
              description: 'desc',
              environmentURL: url,
              target: { host: 'localhost', port: 22 },
              credentials: { type: credType, username: 'user' },
              healthCheck: { enabled: false, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
              commands: { type: 'ssh', restart: { enabled: false } },
              upgradeConfig: { enabled: false, type: 'ssh', upgradeCommand: {} },
            };
            try {
              await onSubmit(testData, 'test-password', 'test-private-key');
            } catch (err) {
              // Swallow rejected promises to prevent unhandled promise rejections
            }
          }}
        >
          {mode === 'create' ? 'Create Environment' : 'Save Environment'}
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
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders create environment heading', () => {
    render(<CreateEnvironment />);
    expect(screen.getByText(/create.*new.*environment/i)).toBeInTheDocument();
  });

  it('renders the environment form', () => {
    render(<CreateEnvironment />);
    expect(screen.getByRole('button', { name: /create environment/i })).toBeInTheDocument();
  });

  it('calls environmentApi.create on form submission with password credentials', async () => {
    mockedEnvApi.create = vi.fn().mockResolvedValue({ id: 'new-env' });
    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });
    render(<CreateEnvironment />);

    await u.type(screen.getByLabelText(/^environment name\s*\*/i), 'Test');
    await u.type(screen.getByLabelText(/^environment url/i), 'https://test.example.com');
    await u.click(screen.getByRole('button', { name: /create environment/i }));

    expect(mockedEnvApi.create).toHaveBeenCalledWith(expect.objectContaining({
      name: 'Test',
      environmentURL: 'https://test.example.com',
      credentials: { type: 'password', username: 'user' },
      metadata: expect.objectContaining({
        password: 'test-password',
      }),
    }));
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });
  });

  it('calls environmentApi.create on form submission with key credentials', async () => {
    mockedEnvApi.create = vi.fn().mockResolvedValue({ id: 'new-env' });
    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });
    render(<CreateEnvironment />);

    await u.type(screen.getByLabelText(/^environment name\s*\*/i), 'Test Key Env');
    await u.selectOptions(screen.getByTestId('cred-type-select'), 'key');
    await u.click(screen.getByRole('button', { name: /create environment/i }));

    expect(mockedEnvApi.create).toHaveBeenCalledWith(expect.objectContaining({
      name: 'Test Key Env',
      credentials: { type: 'key', username: 'user' },
      metadata: expect.objectContaining({
        privateKey: 'test-private-key',
      }),
    }));
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });
  });

  it('displays error message on failed creation', async () => {
    const errorResponse = {
      response: {
        data: {
          message: 'This is a test creation error'
        }
      }
    };
    mockedEnvApi.create = vi.fn().mockRejectedValue(errorResponse);
    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });
    render(<CreateEnvironment />);

    await u.type(screen.getByLabelText(/^environment name\s*\*/i), 'Error Env');
    await u.click(screen.getByRole('button', { name: /create environment/i }));

    await waitFor(() => {
      expect(screen.getByTestId('form-error')).toHaveTextContent('This is a test creation error');
    });
  });

  it('displays fallback error message on failed creation without details', async () => {
    mockedEnvApi.create = vi.fn().mockRejectedValue(new Error('Network error'));
    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });
    render(<CreateEnvironment />);

    await u.type(screen.getByLabelText(/^environment name\s*\*/i), 'Error Env 2');
    await u.click(screen.getByRole('button', { name: /create environment/i }));

    await waitFor(() => {
      expect(screen.getByTestId('form-error')).toHaveTextContent('Failed to create environment');
    });
  });
});

describe('EditEnvironment page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

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

  it('submits update successfully with password metadata', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue({
      ...mockEnvironment,
      credentials: { type: 'password', username: 'user' },
    });
    mockedEnvApi.update = vi.fn().mockResolvedValue({ id: 'env-1' });

    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByText(/edit environment/i)).toBeInTheDocument();
    });

    await u.click(screen.getByRole('button', { name: /save environment/i }));

    expect(mockedEnvApi.update).toHaveBeenCalledWith('env-1', expect.objectContaining({
      name: 'Test Env',
      credentials: { type: 'password', username: 'user' },
      metadata: {
        password: 'test-password',
      },
    }));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });
  });

  it('submits update successfully with key metadata', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue({
      ...mockEnvironment,
      credentials: { type: 'key', username: 'user' },
    });
    mockedEnvApi.update = vi.fn().mockResolvedValue({ id: 'env-1' });

    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByText(/edit environment/i)).toBeInTheDocument();
    });

    await u.selectOptions(screen.getByTestId('cred-type-select'), 'key');
    await u.click(screen.getByRole('button', { name: /save environment/i }));

    expect(mockedEnvApi.update).toHaveBeenCalledWith('env-1', expect.objectContaining({
      credentials: { type: 'key', username: 'user' },
      metadata: {
        privateKey: 'test-private-key',
      },
    }));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });
  });

  it('displays error message on failed update', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockRejectedValue({
      response: {
        data: {
          message: 'Update error has occurred',
        },
      },
    });

    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });
    render(<EditEnvironment />);

    await waitFor(() => {
      expect(screen.getByText(/edit environment/i)).toBeInTheDocument();
    });

    await u.click(screen.getByRole('button', { name: /save environment/i }));

    await waitFor(() => {
      expect(screen.getByTestId('form-error')).toHaveTextContent('Update error has occurred');
    });
  });
});
