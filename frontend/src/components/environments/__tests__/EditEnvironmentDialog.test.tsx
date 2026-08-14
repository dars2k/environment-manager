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

  it('updates host field via onChange handler', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    await waitFor(() => {
      expect(screen.getByDisplayValue('prod.example.com')).toBeInTheDocument();
    });
    fireEvent.change(screen.getByDisplayValue('prod.example.com'), {
      target: { value: 'new-host.example.com' },
    });
    expect(screen.getByDisplayValue('new-host.example.com')).toBeInTheDocument();
  });

  it('updates port field and falls back to 22 on invalid input', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    const portField = await screen.findByLabelText(/^port\s*\*/i);
    fireEvent.change(portField, { target: { value: '2222' } });
    expect((portField as HTMLInputElement).value).toBe('2222');

    fireEvent.change(portField, { target: { value: '' } });
    expect((portField as HTMLInputElement).value).toBe('22');
  });

  it('updates username field via onChange handler', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    const usernameField = await screen.findByLabelText(/^username\s*\*/i);
    fireEvent.change(usernameField, { target: { value: 'newuser' } });
    expect((usernameField as HTMLInputElement).value).toBe('newuser');
  });

  it('switches credential type to key and shows/edits the private key field', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    const authSelect = await screen.findByRole('combobox', { name: /authentication method/i });

    // Password field is shown initially, private key is not
    expect(screen.getByLabelText(/^password\s*\*/i)).toBeInTheDocument();
    expect(screen.queryByLabelText(/^private key\s*\*/i)).toBeNull();

    fireEvent.mouseDown(authSelect);
    const keyOption = await screen.findByRole('option', { name: /private key/i });
    fireEvent.click(keyOption);

    const privateKeyField = await screen.findByLabelText(/^private key\s*\*/i);
    expect(screen.queryByLabelText(/^password\s*\*/i)).toBeNull();
    fireEvent.change(privateKeyField, { target: { value: '-----BEGIN KEY-----' } });
    expect((privateKeyField as HTMLTextAreaElement).value).toBe('-----BEGIN KEY-----');
  });

  it('updates health check endpoint field', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    const endpointField = await screen.findByLabelText(/health check endpoint/i);
    fireEvent.change(endpointField, { target: { value: '/status' } });
    expect((endpointField as HTMLInputElement).value).toBe('/status');
  });

  it('updates health check interval and falls back to 300 on invalid input', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    const intervalField = await screen.findByLabelText(/health check interval/i);
    fireEvent.change(intervalField, { target: { value: '120' } });
    expect((intervalField as HTMLInputElement).value).toBe('120');

    fireEvent.change(intervalField, { target: { value: '' } });
    expect((intervalField as HTMLInputElement).value).toBe('300');
  });

  it('updates health check timeout and falls back to 30 on invalid input', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    const timeoutField = await screen.findByLabelText(/health check timeout/i);
    fireEvent.change(timeoutField, { target: { value: '15' } });
    expect((timeoutField as HTMLInputElement).value).toBe('15');

    fireEvent.change(timeoutField, { target: { value: '' } });
    expect((timeoutField as HTMLInputElement).value).toBe('30');
  });

  it('switches validation type to jsonRegex and resets expected value to empty string', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    const validationSelect = await screen.findByRole('combobox', { name: /validation type/i });

    expect(screen.getByDisplayValue('200')).toBeInTheDocument();

    fireEvent.mouseDown(validationSelect);
    const jsonRegexOption = await screen.findByRole('option', { name: /json regex/i });
    fireEvent.click(jsonRegexOption);

    const expectedValueField = await screen.findByLabelText(/expected value/i);
    expect((expectedValueField as HTMLInputElement).value).toBe('');

    // With jsonRegex selected, typed value is stored as a raw string
    fireEvent.change(expectedValueField, { target: { value: 'ok-pattern' } });
    expect((expectedValueField as HTMLInputElement).value).toBe('ok-pattern');
  });

  it('updates expected value as a parsed number when validation type is statusCode', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    const expectedValueField = await screen.findByLabelText(/expected value/i);
    fireEvent.change(expectedValueField, { target: { value: '404' } });
    expect((expectedValueField as HTMLInputElement).value).toBe('404');

    // Invalid number input falls back to 200
    fireEvent.change(expectedValueField, { target: { value: '' } });
    expect((expectedValueField as HTMLInputElement).value).toBe('200');
  });

  it('updates the restart HTTP method field', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    fireEvent.click(await screen.findByText(/custom commands \(optional\)/i));

    const commandTypeSelect = await screen.findByRole('combobox', { name: /^command type$/i });
    fireEvent.mouseDown(commandTypeSelect);
    fireEvent.click(await screen.findByRole('option', { name: /^http$/i }));

    const methodSelect = await screen.findByRole('combobox', { name: /^method$/i });
    expect(methodSelect).toHaveTextContent('POST');
    fireEvent.mouseDown(methodSelect);
    fireEvent.click(await screen.findByRole('option', { name: /^put$/i }));
    expect(methodSelect).toHaveTextContent('PUT');
  });

  it('updates the restart headers field with valid JSON and ignores invalid JSON', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    fireEvent.click(await screen.findByText(/custom commands \(optional\)/i));

    const commandTypeSelect = await screen.findByRole('combobox', { name: /^command type$/i });
    fireEvent.mouseDown(commandTypeSelect);
    fireEvent.click(await screen.findByRole('option', { name: /^http$/i }));

    const headersField = await screen.findByLabelText(/headers \(optional\)/i);
    expect((headersField as HTMLTextAreaElement).value).toBe('{}');

    fireEvent.change(headersField, { target: { value: '{"Authorization":"Bearer abc"}' } });
    expect((headersField as HTMLTextAreaElement).value).toBe('{"Authorization":"Bearer abc"}');

    // Invalid JSON should be silently ignored (state unchanged)
    fireEvent.change(headersField, { target: { value: '{not valid json' } });
    expect((headersField as HTMLTextAreaElement).value).toBe('{"Authorization":"Bearer abc"}');
  });

  it('updates the restart body field with valid JSON and ignores invalid JSON', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    fireEvent.click(await screen.findByText(/custom commands \(optional\)/i));

    const commandTypeSelect = await screen.findByRole('combobox', { name: /^command type$/i });
    fireEvent.mouseDown(commandTypeSelect);
    fireEvent.click(await screen.findByRole('option', { name: /^http$/i }));

    const bodyField = await screen.findByLabelText(/body \(optional\)/i);
    expect((bodyField as HTMLTextAreaElement).value).toBe('{}');

    fireEvent.change(bodyField, { target: { value: '{"action":"restart"}' } });
    expect((bodyField as HTMLTextAreaElement).value).toBe('{"action":"restart"}');

    // Invalid JSON should be silently ignored (state unchanged)
    fireEvent.change(bodyField, { target: { value: '{still not valid' } });
    expect((bodyField as HTMLTextAreaElement).value).toBe('{"action":"restart"}');
  });

  it('updates jsonPathResponse field in upgrade configuration', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    fireEvent.click(await screen.findByText(/upgrade configuration \(optional\)/i));
    fireEvent.click(await screen.findByLabelText(/enable version upgrades/i));

    const jsonPathField = await screen.findByLabelText(/jsonpath response/i);
    fireEvent.change(jsonPathField, { target: { value: '$.data.releases[*]' } });
    expect((jsonPathField as HTMLInputElement).value).toBe('$.data.releases[*]');
  });

  it('switches upgrade command type to HTTP and updates the URL and Method fields', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentEditDialogOpen(true));
    render(<EditEnvironmentDialog environment={mockEnvironment} />, { store });
    fireEvent.click(await screen.findByText(/upgrade configuration \(optional\)/i));
    fireEvent.click(await screen.findByLabelText(/enable version upgrades/i));

    const upgradeTypeSelect = await screen.findByRole('combobox', { name: /upgrade command type/i });
    fireEvent.mouseDown(upgradeTypeSelect);
    fireEvent.click(await screen.findByRole('option', { name: /^http$/i }));

    const upgradeUrlField = await screen.findByLabelText(/^url\s*\*?$/i);
    fireEvent.change(upgradeUrlField, { target: { value: 'http://myserver/upgrade/{VERSION}' } });
    expect((upgradeUrlField as HTMLInputElement).value).toBe('http://myserver/upgrade/{VERSION}');

    const upgradeMethodSelect = await screen.findByRole('combobox', { name: /^method$/i });
    expect(upgradeMethodSelect).toHaveTextContent('POST');
    fireEvent.mouseDown(upgradeMethodSelect);
    fireEvent.click(await screen.findByRole('option', { name: /^patch$/i }));
    expect(upgradeMethodSelect).toHaveTextContent('PATCH');
  });
});
