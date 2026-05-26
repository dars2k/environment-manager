import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@/test/test-utils';
import { EditEnvironment } from '../EditEnvironment';
import { environmentApi } from '@/api/environments';
import { HealthStatus } from '@/types/environment';
import userEvent from '@testing-library/user-event';

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

describe('EditEnvironment page coverage', () => {
  it('submits the update form', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockResolvedValue({ ...mockEnvironment, name: 'Updated' });

    render(<EditEnvironment />);
    const user = userEvent.setup();

    await waitFor(() => expect(screen.getByText('Edit Environment')).toBeInTheDocument());

    const nameInput = screen.getByLabelText(/environment name/i);
    await user.clear(nameInput);
    await user.type(nameInput, 'Updated Env');

    const submitButton = screen.getByRole('button', { name: /update environment/i });
    await user.click(submitButton);

    await waitFor(() => expect(mockedEnvApi.update).toHaveBeenCalled());
  });

  it('shows error on failed update', async () => {
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
    mockedEnvApi.update = vi.fn().mockRejectedValue({
      response: { data: { message: 'Update failed' } }
    });

    render(<EditEnvironment />);
    const user = userEvent.setup();

    await waitFor(() => expect(screen.getByText('Edit Environment')).toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: /update environment/i }));

    await waitFor(() => expect(screen.getByText('Update failed')).toBeInTheDocument());
  });

  it('submits with password', async () => {
    const keyEnv = { ...mockEnvironment, credentials: { type: 'key', username: 'user' } };
    mockedEnvApi.get = vi.fn().mockResolvedValue(keyEnv);
    mockedEnvApi.update = vi.fn().mockResolvedValue(keyEnv);

    render(<EditEnvironment />);
    const user = userEvent.setup();

    await waitFor(() => expect(screen.getByText('Edit Environment')).toBeInTheDocument());

    await user.click(screen.getByRole('button', { name: /update environment/i }));
    await waitFor(() => expect(mockedEnvApi.update).toHaveBeenCalled());
  });
});
