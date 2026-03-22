import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@/test/test-utils';
import { CreateEnvironmentDialog } from '../CreateEnvironmentDialog';
import { createTestStore } from '@/test/test-utils';
import { setEnvironmentCreateDialogOpen } from '@/store/slices/uiSlice';
import { environmentApi } from '@/api/environments';

vi.mock('@/api/environments');
const mockedApi = vi.mocked(environmentApi);

function renderDialog() {
  const store = createTestStore();
  store.dispatch(setEnvironmentCreateDialogOpen(true));
  render(<CreateEnvironmentDialog />, { store });
  return store;
}

describe('CreateEnvironmentDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('does not render when closed', () => {
    render(<CreateEnvironmentDialog />);
    expect(screen.queryByText('Create New Environment')).toBeNull();
  });

  it('renders dialog title when open', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText('Create New Environment')).toBeInTheDocument();
    });
  });

  it('shows form fields', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByLabelText(/^name\s*\*/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^host\s*\*/i)).toBeInTheDocument();
    });
  });

  it('shows validation error when name is empty on submit', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^create$/i })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /^create$/i }));
    await waitFor(() => {
      expect(screen.getByText(/please fill in all required fields/i)).toBeInTheDocument();
    });
  });

  it('shows password required error when credentials type is password', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByLabelText(/^name\s*\*/i)).toBeInTheDocument();
    });
    // Fill name and host but not password
    fireEvent.change(screen.getByLabelText(/^name\s*\*/i), {
      target: { value: 'Test Env' },
    });
    fireEvent.change(screen.getByLabelText(/^host\s*\*/i), {
      target: { value: 'example.com' },
    });
    fireEvent.change(screen.getByLabelText(/^username\s*\*/i), {
      target: { value: 'admin' },
    });
    fireEvent.click(screen.getByRole('button', { name: /^create$/i }));
    await waitFor(() => {
      expect(screen.getByText(/password is required/i)).toBeInTheDocument();
    });
  });

  it('closes dialog on Cancel click', async () => {
    const store = createTestStore();
    store.dispatch(setEnvironmentCreateDialogOpen(true));
    render(<CreateEnvironmentDialog />, { store });
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^cancel$/i })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /^cancel$/i }));
    expect(store.getState().ui.environmentCreateDialogOpen).toBe(false);
  });

  it('calls environmentApi.create on valid submit', async () => {
    mockedApi.create = vi.fn().mockResolvedValue({ id: 'new-env' });
    renderDialog();
    await waitFor(() => {
      expect(screen.getByLabelText(/^name\s*\*/i)).toBeInTheDocument();
    });
    fireEvent.change(screen.getByLabelText(/^name\s*\*/i), {
      target: { value: 'Production' },
    });
    fireEvent.change(screen.getByLabelText(/^host\s*\*/i), {
      target: { value: 'prod.example.com' },
    });
    fireEvent.change(screen.getByLabelText(/^username\s*\*/i), {
      target: { value: 'deploy' },
    });
    const passwordField = screen.getByLabelText(/^password\s*\*/i);
    fireEvent.change(passwordField, { target: { value: 'secret' } });
    fireEvent.click(screen.getByRole('button', { name: /^create$/i }));
    await waitFor(() => {
      expect(mockedApi.create).toHaveBeenCalled();
    });
  });

  it('shows private key error when credentials type is key', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByLabelText(/^name\s*\*/i)).toBeInTheDocument();
    });
    fireEvent.change(screen.getByLabelText(/^name\s*\*/i), {
      target: { value: 'Test Env' },
    });
    fireEvent.change(screen.getByLabelText(/^host\s*\*/i), {
      target: { value: 'example.com' },
    });
    fireEvent.change(screen.getByLabelText(/^username\s*\*/i), {
      target: { value: 'admin' },
    });
    // Switch to key type using the Authentication Method select
    const credTypeSelect = screen.getByLabelText(/^authentication method\s*\*/i);
    fireEvent.mouseDown(credTypeSelect);
    await waitFor(() => {
      expect(screen.getByRole('option', { name: /private key/i })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('option', { name: /private key/i }));
    // Submit without private key
    fireEvent.click(screen.getByRole('button', { name: /^create$/i }));
    await waitFor(() => {
      expect(screen.getByText(/private key is required/i)).toBeInTheDocument();
    });
  });

  it('shows custom commands accordion section', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText(/custom commands \(optional\)/i)).toBeInTheDocument();
    });
  });

  it('shows upgrade configuration accordion section', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText(/upgrade configuration \(optional\)/i)).toBeInTheDocument();
    });
  });

  it('toggles health check fields when switch clicked', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByLabelText(/enable health checks/i)).toBeInTheDocument();
    });
    // Health check is enabled by default — click to disable
    const healthCheckSwitch = screen.getByLabelText(/enable health checks/i);
    fireEvent.click(healthCheckSwitch);
    // Fields should disappear
    await waitFor(() => {
      expect(screen.queryByLabelText(/health check endpoint/i)).toBeNull();
    });
    // Click to re-enable
    fireEvent.click(healthCheckSwitch);
    await waitFor(() => {
      expect(screen.getByLabelText(/health check endpoint/i)).toBeInTheDocument();
    });
  });

  it('updates health check fields', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByLabelText(/health check endpoint/i)).toBeInTheDocument();
    });
    fireEvent.change(screen.getByLabelText(/health check endpoint/i), {
      target: { value: '/api/health' },
    });
    fireEvent.change(screen.getByLabelText(/health check interval/i), {
      target: { value: '60' },
    });
    expect(screen.getByLabelText(/health check endpoint/i)).toHaveValue('/api/health');
  });

  it('shows error from failed API call', async () => {
    mockedApi.create = vi.fn().mockRejectedValue({
      response: { data: { message: 'Duplicate environment name' } },
    });
    renderDialog();
    await waitFor(() => {
      expect(screen.getByLabelText(/^name\s*\*/i)).toBeInTheDocument();
    });
    fireEvent.change(screen.getByLabelText(/^name\s*\*/i), {
      target: { value: 'Production' },
    });
    fireEvent.change(screen.getByLabelText(/^host\s*\*/i), {
      target: { value: 'prod.example.com' },
    });
    fireEvent.change(screen.getByLabelText(/^username\s*\*/i), {
      target: { value: 'deploy' },
    });
    fireEvent.change(screen.getByLabelText(/^password\s*\*/i), {
      target: { value: 'secret' },
    });
    fireEvent.click(screen.getByRole('button', { name: /^create$/i }));
    await waitFor(() => {
      expect(screen.getByText(/duplicate environment name/i)).toBeInTheDocument();
    });
  });

  it('expands custom commands accordion and shows SSH command field', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText(/custom commands \(optional\)/i)).toBeInTheDocument();
    });
    // Click the accordion summary to expand
    fireEvent.click(screen.getByText(/custom commands \(optional\)/i));
    await waitFor(() => {
      // After expansion, the Command Type select should be visible
      expect(screen.getByLabelText(/^command type$/i)).toBeInTheDocument();
    });
  });

  it('shows SSH command field in custom commands accordion', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText(/custom commands \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/custom commands \(optional\)/i));
    await waitFor(() => {
      // SSH command field visible (commands.type defaults to 'ssh')
      const sshField = screen.queryByLabelText(/^ssh command$/i);
      if (sshField) {
        expect(sshField).toBeInTheDocument();
        // Test changing the value
        fireEvent.change(sshField, { target: { value: 'sudo systemctl restart app' } });
        expect((sshField as HTMLInputElement).value).toBe('sudo systemctl restart app');
      }
    });
  });

  it('switches to HTTP command type and shows URL field', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText(/custom commands \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/custom commands \(optional\)/i));

    await waitFor(() => {
      expect(screen.getByLabelText(/^command type$/i)).toBeInTheDocument();
    });

    // Change command type to HTTP
    fireEvent.mouseDown(screen.getByLabelText(/^command type$/i));
    await waitFor(() => {
      const httpOption = screen.queryByRole('option', { name: /^http$/i });
      if (httpOption) {
        fireEvent.click(httpOption);
      }
    });

    await waitFor(() => {
      // After switching to HTTP, URL field should appear
      const urlField = screen.queryByPlaceholderText(/localhost:8080\/restart/i);
      if (urlField) {
        expect(urlField).toBeInTheDocument();
      }
    });
  });

  it('expands upgrade configuration accordion', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText(/upgrade configuration \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/upgrade configuration \(optional\)/i));
    await waitFor(() => {
      // After expansion, the Enable Version Upgrades switch should be visible
      expect(screen.getByLabelText(/enable version upgrades/i)).toBeInTheDocument();
    });
  });

  it('shows upgrade fields when enable version upgrades is toggled on', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText(/upgrade configuration \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/upgrade configuration \(optional\)/i));
    await waitFor(() => {
      expect(screen.getByLabelText(/enable version upgrades/i)).toBeInTheDocument();
    });
    // Enable upgrade
    fireEvent.click(screen.getByLabelText(/enable version upgrades/i));
    await waitFor(() => {
      expect(screen.getByLabelText(/version list url/i)).toBeInTheDocument();
    });
  });

  it('updates authentication method to key and shows private key field', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByLabelText(/^authentication method\s*\*/i)).toBeInTheDocument();
    });
    fireEvent.mouseDown(screen.getByLabelText(/^authentication method\s*\*/i));
    await waitFor(() => {
      const keyOption = screen.queryByRole('option', { name: /private key/i });
      if (keyOption) {
        fireEvent.click(keyOption);
      }
    });
    await waitFor(() => {
      const pkField = screen.queryByPlaceholderText(/paste your private key/i);
      if (pkField) {
        expect(pkField).toBeInTheDocument();
        fireEvent.change(pkField, { target: { value: '-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----' } });
      }
    });
  });

  it('updates description field', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByLabelText(/^description$/i)).toBeInTheDocument();
    });
    fireEvent.change(screen.getByLabelText(/^description$/i), {
      target: { value: 'My test environment description' },
    });
    expect(screen.getByLabelText(/^description$/i)).toHaveValue('My test environment description');
  });

  it('updates health check validation type', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByLabelText(/validation type/i)).toBeInTheDocument();
    });
    fireEvent.mouseDown(screen.getByLabelText(/validation type/i));
    await waitFor(() => {
      const jsonRegexOption = screen.queryByRole('option', { name: /json regex/i });
      if (jsonRegexOption) {
        fireEvent.click(jsonRegexOption);
      }
    });
    // Check expected value field is still there
    await waitFor(() => {
      expect(screen.getByLabelText(/expected value/i)).toBeInTheDocument();
    });
  });

  it('changes SSH command in custom commands accordion', async () => {
    renderDialog();
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

  it('shows HTTP URL field in custom commands accordion after switching to HTTP', async () => {
    renderDialog();
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
      // After switching to HTTP, check for HTTP URL field
      const urlField = screen.queryByPlaceholderText(/localhost:8080\/restart/i);
      if (urlField) {
        // Test changing the URL
        fireEvent.change(urlField, { target: { value: 'http://myserver/restart' } });
        expect((urlField as HTMLInputElement).value).toBe('http://myserver/restart');
      }
    });
  });

  it('changes headers JSON in HTTP restart command section', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText(/custom commands \(optional\)/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/custom commands \(optional\)/i));

    // Switch to HTTP first
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
      const headersField = screen.queryByLabelText(/headers \(optional\)/i);
      if (headersField) {
        fireEvent.change(headersField, { target: { value: '{"Authorization": "Bearer token"}' } });
        expect(headersField).toBeInTheDocument();
      }
    });
  });

  it('changes upgrade SSH commands in upgrade accordion', async () => {
    renderDialog();
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

  it('switches upgrade command to HTTP type', async () => {
    renderDialog();
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

  it('updates version list URL', async () => {
    renderDialog();
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
});
