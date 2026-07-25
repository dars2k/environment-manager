import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@/test/test-utils';
import { EditEnvironment } from '../EditEnvironment';
import { environmentApi } from '@/api/environments';
import { HealthStatus } from '@/types/environment';

const mockEnqueueSnackbar = vi.fn();
vi.mock('notistack', () => ({
  useSnackbar: () => ({
    enqueueSnackbar: mockEnqueueSnackbar,
    closeSnackbar: vi.fn(),
  }),
}));

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ id: 'env-1' }),
  };
});

vi.mock('@/api/environments');
const mockedEnvApi = vi.mocked(environmentApi);

let capturedOnSubmit: ((data: any, password?: string, privateKey?: string) => Promise<void>) | undefined;
let capturedError: string | null = null;

vi.mock('@/components/environments/EnvironmentForm', () => ({
  EnvironmentForm: (props: any) => {
    capturedOnSubmit = props.onSubmit;
    capturedError = props.error;
    return (
      <div>
        {props.error && <div data-testid="form-error">{props.error}</div>}
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

const baseSubmitData = {
  name: 'Updated Env',
  description: 'updated desc',
  environmentURL: 'https://updated.example.com',
  target: { host: 'localhost', port: 22 },
  credentials: { type: 'password', username: 'user' },
  healthCheck: mockEnvironment.healthCheck,
  commands: mockEnvironment.commands,
  upgradeConfig: mockEnvironment.upgradeConfig,
};

describe('EditEnvironment page submit flow', () => {
  beforeEach(() => {
    capturedOnSubmit = undefined;
    capturedError = null;
    mockNavigate.mockClear();
    mockEnqueueSnackbar.mockClear();
    mockedEnvApi.get = vi.fn().mockResolvedValue(mockEnvironment);
  });

  it('updates the environment with a password and navigates on success', async () => {
    mockedEnvApi.update = vi.fn().mockResolvedValue({ ...mockEnvironment, name: 'Updated Env' });

    render(<EditEnvironment />);
    await waitFor(() => expect(capturedOnSubmit).toBeDefined());

    await capturedOnSubmit!(baseSubmitData, 'super-secret');

    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalledWith(
        'env-1',
        expect.objectContaining({
          name: 'Updated Env',
          metadata: { password: 'super-secret' },
        })
      );
    });
    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/dashboard'));
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
      'Environment updated successfully',
      expect.objectContaining({ variant: 'success' })
    );
  });

  it('updates the environment with a private key when credentials type is key', async () => {
    mockedEnvApi.update = vi.fn().mockResolvedValue(mockEnvironment);

    render(<EditEnvironment />);
    await waitFor(() => expect(capturedOnSubmit).toBeDefined());

    const keyData = { ...baseSubmitData, credentials: { type: 'key', username: 'user' } };
    await capturedOnSubmit!(keyData, undefined, 'private-key-contents');

    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalledWith(
        'env-1',
        expect.objectContaining({ metadata: { privateKey: 'private-key-contents' } })
      );
    });
  });

  it('omits metadata when no password or private key is provided', async () => {
    mockedEnvApi.update = vi.fn().mockResolvedValue(mockEnvironment);

    render(<EditEnvironment />);
    await waitFor(() => expect(capturedOnSubmit).toBeDefined());

    await capturedOnSubmit!(baseSubmitData);

    await waitFor(() => {
      expect(mockedEnvApi.update).toHaveBeenCalledWith(
        'env-1',
        expect.not.objectContaining({ metadata: expect.anything() })
      );
    });
  });

  it('shows an error message when the update fails with a server message', async () => {
    mockedEnvApi.update = vi.fn().mockRejectedValue({
      response: { data: { message: 'Name already in use' } },
    });

    render(<EditEnvironment />);
    await waitFor(() => expect(capturedOnSubmit).toBeDefined());

    await capturedOnSubmit!(baseSubmitData).catch(() => {});

    await waitFor(() => expect(capturedError).toBe('Name already in use'));
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('falls back to a generic error message when the server provides none', async () => {
    mockedEnvApi.update = vi.fn().mockRejectedValue(new Error('network down'));

    render(<EditEnvironment />);
    await waitFor(() => expect(capturedOnSubmit).toBeDefined());

    await capturedOnSubmit!(baseSubmitData).catch(() => {});

    await waitFor(() => expect(capturedError).toBe('Failed to update environment'));
  });
});
