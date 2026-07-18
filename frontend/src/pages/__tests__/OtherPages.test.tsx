import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
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

// Mock EnvironmentForm to allow direct triggering of onSubmit
vi.mock('@/components/environments/EnvironmentForm', () => ({
  EnvironmentForm: ({ onSubmit, error }: any) => {
    const handleSafeSubmit = async (data: any, password?: string, privateKey?: string) => {
      try {
        await onSubmit(data, password, privateKey);
      } catch (e) {}
    };
    return (
      <div>
        <div data-testid="form-error">{error}</div>
        <button data-testid="submit-form" onClick={() => handleSafeSubmit({ name: 'Form Env', target: { host: 'localhost', port: 22 }, credentials: { type: 'password', username: 'user' } })}>
          Submit Form
        </button>
        <button data-testid="submit-form-key" onClick={() => handleSafeSubmit({ name: 'Form Env', target: { host: 'localhost', port: 22 }, credentials: { type: 'key', username: 'user' } }, undefined, 'pk')}>
          Submit Form Key
        </button>
        <button data-testid="submit-form-pass" onClick={() => handleSafeSubmit({ name: 'Form Env', target: { host: 'localhost', port: 22 }, credentials: { type: 'password', username: 'user' } }, 'pass')}>
          Submit Form Pass
        </button>
        <button>Create Environment</button>
      </div>
    );
  },
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
    expect(screen.getByRole('button', { name: /create environment/i })).toBeInTheDocument();
  });

  it('calls environmentApi.create on form submission success', async () => {
    mockedEnvApi.create = vi.fn().mockResolvedValue({ id: 'new-env' });
    render(<CreateEnvironment />);

    // Trigger submit with password
    fireEvent.click(screen.getByTestId('submit-form-pass'));
    await waitFor(() => {
      expect(mockedEnvApi.create).toHaveBeenCalled();
    });

    // Trigger submit with key
    fireEvent.click(screen.getByTestId('submit-form-key'));
    await waitFor(() => {
      expect(mockedEnvApi.create).toHaveBeenCalledTimes(2);
    });
  });

  it('handles environmentApi.create error', async () => {
    mockedEnvApi.create = vi.fn().mockRejectedValue({
      response: { data: { message: 'Database creation failed' } },
    });
    render(<CreateEnvironment />);

    fireEvent.click(screen.getByTestId('submit-form'));
    await waitFor(() => {
      expect(screen.getByTestId('form-error')).toHaveTextContent('Database creation failed');
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

  it('calls environmentApi.update on form submission', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockResolvedValue({ ...mockEnvironment, id: 'env-1' });
    render(<EditEnvironment />);
    await waitFor(() => {
      expect(screen.getByText(/edit environment/i)).toBeInTheDocument();
    });

    // Submit form with password update
    fireEvent.click(screen.getByTestId('submit-form-pass'));
    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalled();
    });

    // Submit form with key update
    fireEvent.click(screen.getByTestId('submit-form-key'));
    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalledTimes(2);
    });
  });

  it('handles edit environment update error', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockRejectedValue({
      response: { data: { message: 'Database update failed' } },
    });
    render(<EditEnvironment />);
    await waitFor(() => {
      expect(screen.getByText(/edit environment/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('submit-form'));
    await waitFor(() => {
      expect(screen.getByTestId('form-error')).toHaveTextContent('Database update failed');
    });
  });
});
