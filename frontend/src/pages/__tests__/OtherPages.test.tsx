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

// Mock EnvironmentForm component and wrap the asynchronous submission handler in try-catch to swallow rejected promises
vi.mock('@/components/environments/EnvironmentForm', () => ({
  EnvironmentForm: ({ onSubmit, error }: any) => (
    <div>
      {error && <div>{error}</div>}
      <button
        onClick={async () => {
          try {
            await onSubmit(mockEnvironment, 'new-password', 'new-key');
          } catch (e) {
            // Swallow rejected promises to prevent unhandled promise rejections in tests
          }
        }}
      >
        Submit Form Mock
      </button>
    </div>
  ),
}));

const mockEnvironment = {
  id: 'env-1',
  name: 'Test Env',
  description: 'desc',
  environmentURL: 'https://test.example.com',
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
    expect(screen.getByRole('button', { name: /Submit Form Mock/i })).toBeInTheDocument();
  });

  it('calls environmentApi.create on form submission', async () => {
    mockedEnvApi.create = vi.fn().mockResolvedValue({ id: 'new-env' });
    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });
    render(<CreateEnvironment />);
    await u.click(screen.getByRole('button', { name: /Submit Form Mock/i }));
    expect(mockedEnvApi.create).toHaveBeenCalled();
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

  it('handles successful update on form submission with password metadata', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockResolvedValue({ id: 'env-1' });
    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });

    render(<EditEnvironment />);
    await waitFor(() => {
      expect(screen.getByText(/edit environment/i)).toBeInTheDocument();
    });

    // Submit the form
    const saveButton = screen.getByRole('button', { name: /Submit Form Mock/i });
    await u.click(saveButton);

    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalled();
    });
  });

  it('handles failed update and displays error alert', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockImplementation(() => Promise.reject({
      response: { data: { message: 'Some API Error' } }
    }));
    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });

    render(<EditEnvironment />);
    await waitFor(() => {
      expect(screen.getByText(/edit environment/i)).toBeInTheDocument();
    });

    const saveButton = screen.getByRole('button', { name: /Submit Form Mock/i });
    await u.click(saveButton);

    await waitFor(() => {
      expect(screen.getByText('Some API Error')).toBeInTheDocument();
    });
  });

  it('handles failed update with fallback generic error', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockImplementation(() => Promise.reject(new Error('Generic Error')));
    const user = (await import('@testing-library/user-event')).default;
    const u = user.setup({ delay: null });

    render(<EditEnvironment />);
    await waitFor(() => {
      expect(screen.getByText(/edit environment/i)).toBeInTheDocument();
    });

    const saveButton = screen.getByRole('button', { name: /Submit Form Mock/i });
    await u.click(saveButton);

    await waitFor(() => {
      expect(screen.getByText('Failed to update environment')).toBeInTheDocument();
    });
  });
});
