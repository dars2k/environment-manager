import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
import { NotFound } from '../NotFound';
import { CreateEnvironment } from '../CreateEnvironment';
import { EditEnvironment } from '../EditEnvironment';
import { environmentApi } from '@/api/environments';
import { HealthStatus } from '@/types/environment';

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

describe('EditEnvironment page', () => {
  it('shows error when update fails', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockRejectedValue({
      response: { data: { message: 'Update failed' } }
    });

    render(<EditEnvironment />);
    await waitFor(() => expect(screen.getByText(/test env/i)).toBeInTheDocument());

    const saveBtn = screen.getByRole('button', { name: /update environment/i });
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(screen.getByText('Update failed')).toBeInTheDocument();
    });
  });

  it('navigates to dashboard on successful update', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockResolvedValue({ ...mockEnvironment, name: 'Updated' });

    render(<EditEnvironment />);
    await waitFor(() => expect(screen.getByText(/test env/i)).toBeInTheDocument());

    const saveBtn = screen.getByRole('button', { name: /update environment/i });
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });
  });
});
