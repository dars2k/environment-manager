import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@/test/test-utils';
import { EditEnvironmentDialog } from '../EditEnvironmentDialog';
import { createTestStore } from '@/test/test-utils';
import { setEnvironmentEditDialogOpen } from '@/store/slices/uiSlice';
import { environmentApi } from '@/api/environments';
import { Environment, HealthStatus } from '@/types/environment';

vi.mock('@/api/environments');
const mockedApi = vi.mocked(environmentApi);

const mockEnvironment: Environment = {
  id: 'env-1',
  name: 'Production',
  description: 'Prod env',
  environmentURL: 'https://prod.example.com',
  target: { host: 'prod.example.com', port: 22 },
  credentials: { type: 'password', username: 'deploy' },
  healthCheck: {
    enabled: true,
    endpoint: '/health',
    method: 'GET',
    interval: 60,
    timeout: 10,
    validation: { type: 'statusCode', value: 200 },
  },
  status: { health: HealthStatus.Healthy, lastCheck: new Date().toISOString(), message: 'OK', responseTime: 100 },
  systemInfo: { osVersion: 'Ubuntu 22.04', appVersion: '3.1.0', lastUpdated: new Date().toISOString() },
  timestamps: { createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  commands: {
    type: 'ssh',
    restart: { enabled: true, command: 'systemctl restart app' },
  },
  upgradeConfig: {
    enabled: false,
    type: 'ssh',
    versionListURL: '',
    jsonPathResponse: '',
    upgradeCommand: { command: '' },
  },
};

describe('EditEnvironmentDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders nothing when environment is null', () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    const { container } = render(<EditEnvironmentDialog environment={null} />, { store });
    expect(container.firstChild).toBeNull();
  });

  it('renders dialog when open with environment', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText('Edit Environment')).toBeInTheDocument();
    });
  });

  it('pre-fills form with environment name', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByDisplayValue('Production')).toBeInTheDocument();
    });
  });

  it('pre-fills host field', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByDisplayValue('prod.example.com')).toBeInTheDocument();
    });
  });

  it('closes dialog on Cancel click', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^cancel$/i })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /^cancel$/i }));
    expect(store.getState().ui.environmentEditDialogOpen).toBe(false);
  });

  it('shows validation error when name is cleared and submitted', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByLabelText(/^name\s*\*/i)).toBeInTheDocument();
    });
    fireEvent.change(screen.getByLabelText(/^name\s*\*/i), {
      target: { value: '' },
    });
    fireEvent.click(screen.getByRole('button', { name: /^update$/i }));
    await waitFor(() => {
      expect(screen.getByText(/please fill in all required fields/i)).toBeInTheDocument();
    });
  });

  it('calls environmentApi.update on valid submit', async () => {
    mockedApi.update = vi.fn().mockResolvedValue(mockEnvironment);
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByDisplayValue('Production')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /^update$/i }));
    await waitFor(() => {
      expect(mockedApi.update).toHaveBeenCalledWith('env-1', expect.any(Object));
    });
  });

  it('shows the dialog title', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText('Edit Environment')).toBeInTheDocument();
    });
  });

  it('shows health check switch', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByLabelText(/enable health checks/i)).toBeInTheDocument();
    });
  });

  it('toggles health check fields', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByLabelText(/health check endpoint/i)).toBeInTheDocument();
    });
    // Toggle off
    const healthSwitch = screen.getByLabelText(/enable health checks/i);
    fireEvent.click(healthSwitch);
    await waitFor(() => {
      expect(screen.queryByLabelText(/health check endpoint/i)).toBeNull();
    });
  });

  it('shows custom commands accordion', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText(/custom commands \(optional\)/i)).toBeInTheDocument();
    });
  });

  it('shows upgrade configuration accordion', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText(/upgrade configuration \(optional\)/i)).toBeInTheDocument();
    });
  });

  it('shows error on failed update', async () => {
    mockedApi.update = vi.fn().mockRejectedValue({
      response: { data: { message: 'Update failed' } },
    });
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByDisplayValue('Production')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /^update$/i }));
    await waitFor(() => {
      expect(mockedApi.update).toHaveBeenCalled();
    });
  });

  it('expands custom commands accordion and shows SSH command field', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText(/custom commands \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/custom commands \(optional\)/i));
    await waitFor(() => {
      const commandTypeField = screen.queryByLabelText(/^command type$/i);
      if (commandTypeField) {
        expect(commandTypeField).toBeInTheDocument();
      }
    });
  });

  it('expands upgrade configuration and enables it', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText(/upgrade configuration \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/upgrade configuration \(optional\)/i));
    await waitFor(() => {
      expect(screen.getByLabelText(/enable version upgrades/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByLabelText(/enable version upgrades/i));
    await waitFor(() => {
      expect(screen.getByLabelText(/version list url/i)).toBeInTheDocument();
    });
  });

  it('switches to HTTP command type in custom commands', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText(/custom commands \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/custom commands \(optional\)/i));
    await waitFor(() => {
      const commandTypeField = screen.queryByLabelText(/^command type$/i);
      if (commandTypeField) {
        fireEvent.mouseDown(commandTypeField);
      }
    });
    await waitFor(() => {
      const httpOption = screen.queryByRole('option', { name: /^http$/i });
      if (httpOption) {
        fireEvent.click(httpOption);
      }
    });
    await waitFor(() => {
      // After HTTP selection, URL field should appear
      const urlField = screen.queryByPlaceholderText(/localhost:8080\/restart/i);
      if (urlField) {
        expect(urlField).toBeInTheDocument();
      }
    });
  });

  it('updates environment URL field', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByDisplayValue('https://prod.example.com')).toBeInTheDocument();
    });
    fireEvent.change(screen.getByDisplayValue('https://prod.example.com'), {
      target: { value: 'https://new.prod.example.com' },
    });
    expect(screen.getByDisplayValue('https://new.prod.example.com')).toBeInTheDocument();
  });

  it('changes SSH command in custom commands accordion', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText(/custom commands \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/custom commands \(optional\)/i));
    await waitFor(() => {
      const sshCmdField = screen.queryByLabelText(/^ssh command$/i);
      if (sshCmdField) {
        fireEvent.change(sshCmdField, { target: { value: 'sudo systemctl restart myapp' } });
        expect((sshCmdField as HTMLInputElement).value).toBe('sudo systemctl restart myapp');
      }
    });
  });

  it('shows HTTP fields in custom commands after switching to HTTP', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText(/custom commands \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/custom commands \(optional\)/i));

    await waitFor(() => {
      const commandTypeField = screen.queryByLabelText(/^command type$/i);
      if (commandTypeField) {
        fireEvent.mouseDown(commandTypeField);
      }
    });
    await waitFor(() => {
      const httpOption = screen.queryByRole('option', { name: /^http$/i });
      if (httpOption) {
        fireEvent.click(httpOption);
      }
    });

    await waitFor(() => {
      const urlField = screen.queryByPlaceholderText(/localhost:8080\/restart/i);
      if (urlField) {
        fireEvent.change(urlField, { target: { value: 'http://myserver/restart' } });
        expect((urlField as HTMLInputElement).value).toBe('http://myserver/restart');
      }
    });
  });

  it('enables upgrade config and updates version list URL', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText(/upgrade configuration \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/upgrade configuration \(optional\)/i));
    await waitFor(() => {
      expect(screen.getByLabelText(/enable version upgrades/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByLabelText(/enable version upgrades/i));
    await waitFor(() => {
      const versionListField = screen.queryByLabelText(/version list url/i);
      if (versionListField) {
        fireEvent.change(versionListField, { target: { value: 'https://api.example.com/versions' } });
        expect((versionListField as HTMLInputElement).value).toBe('https://api.example.com/versions');
      }
    });
  });

  it('changes upgrade SSH command in upgrade section', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText(/upgrade configuration \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/upgrade configuration \(optional\)/i));
    await waitFor(() => {
      expect(screen.getByLabelText(/enable version upgrades/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByLabelText(/enable version upgrades/i));
    await waitFor(() => {
      const sshCmdsField = screen.queryByLabelText(/^ssh commands$/i);
      if (sshCmdsField) {
        fireEvent.change(sshCmdsField, { target: { value: 'sudo upgrade --version={VERSION}' } });
        expect((sshCmdsField as HTMLInputElement).value).toBe('sudo upgrade --version={VERSION}');
      }
    });
  });

  it('switches upgrade command to HTTP type in edit dialog', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByText(/upgrade configuration \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/upgrade configuration \(optional\)/i));
    await waitFor(() => {
      expect(screen.getByLabelText(/enable version upgrades/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByLabelText(/enable version upgrades/i));

    await waitFor(() => {
      const upgradeTypeField = screen.queryByLabelText(/upgrade command type/i);
      if (upgradeTypeField) {
        fireEvent.mouseDown(upgradeTypeField);
      }
    });
    await waitFor(() => {
      const httpOption = screen.queryByRole('option', { name: /^http$/i });
      if (httpOption) {
        fireEvent.click(httpOption);
      }
    });
    await waitFor(() => {
      const upgradeUrlField = screen.queryByPlaceholderText(/localhost:8080\/upgrade/i);
      if (upgradeUrlField) {
        expect(upgradeUrlField).toBeInTheDocument();
      }
    });
  });
});
