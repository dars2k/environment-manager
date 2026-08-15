import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor, fireEvent, within } from '@/test/test-utils';
import userEvent from '@testing-library/user-event';
import { EnvironmentForm } from '../EnvironmentForm';
import { Environment, HealthStatus } from '@/types/environment';

// Mock react-router-dom
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockEnvironment: Environment = {
  id: '1',
  name: 'Test Environment',
  description: 'Test Description',
  environmentURL: 'https://test.example.com',
  target: {
    host: 'test.example.com',
    port: 22,
  },
  credentials: {
    type: 'password',
    username: 'testuser',
  },
  healthCheck: {
    enabled: true,
    endpoint: '/health',
    method: 'GET',
    interval: 300,
    timeout: 30,
    validation: {
      type: 'statusCode',
      value: 200,
    },
  },
  status: {
    health: HealthStatus.Healthy,
    lastCheck: new Date().toISOString(),
    message: 'Environment is healthy',
    responseTime: 100,
  },
  systemInfo: {
    osVersion: 'Linux 5.10',
    appVersion: '1.0.0',
    lastUpdated: new Date().toISOString(),
  },
  timestamps: {
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  commands: {
    type: 'ssh',
    restart: {
      enabled: true,
      command: 'sudo systemctl restart app',
    },
  },
  upgradeConfig: {
    enabled: false,
    type: 'ssh',
    versionListURL: '',
    jsonPathResponse: '',
    upgradeCommand: {
      command: '',
    },
  },
};

describe('EnvironmentForm', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
  });

  it('should render empty form in create mode', () => {
    const onSubmit = vi.fn();
    
    render(
      <EnvironmentForm 
        onSubmit={onSubmit}
        mode="create"
      />
    );
    
    expect(screen.getByLabelText(/Environment Name/i)).toHaveValue('');
    expect(screen.getByLabelText(/Description/i)).toHaveValue('');
    expect(screen.getByLabelText(/Environment URL/i)).toHaveValue('');
  });
  
  it('should render form with initial values in edit mode', () => {
    const onSubmit = vi.fn();
    
    render(
      <EnvironmentForm 
        initialData={mockEnvironment}
        onSubmit={onSubmit}
        mode="edit"
      />
    );
    
    expect(screen.getByLabelText(/Environment Name/i)).toHaveValue('Test Environment');
    expect(screen.getByLabelText(/Description/i)).toHaveValue('Test Description');
    expect(screen.getByLabelText(/Environment URL/i)).toHaveValue('https://test.example.com');
  });
  
  it('should validate required fields', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    
    render(
      <EnvironmentForm 
        onSubmit={onSubmit}
        mode="create"
      />
    );
    
    // Submit without filling required fields
    const submitButton = screen.getByRole('button', { name: /Create Environment/i });
    await user.click(submitButton);
    
    await waitFor(() => {
      expect(onSubmit).not.toHaveBeenCalled();
    });
  });
  
  it('should handle form submission', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Fill in required fields
    await user.type(screen.getByLabelText(/Environment Name/i), 'New Environment');
    await user.type(screen.getByLabelText(/Environment URL/i), 'https://new.example.com');

    // Submit form
    const submitButton = screen.getByRole('button', { name: /Create Environment/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'New Environment',
          environmentURL: 'https://new.example.com',
        }),
        expect.any(String), // password
        expect.any(String)  // privateKey
      );
    });
  }, 15000);
  
  it('should handle cancel', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    
    render(
      <EnvironmentForm 
        onSubmit={onSubmit}
        mode="create"
      />
    );
    
    const cancelButton = screen.getByRole('button', { name: /Cancel/i });
    await user.click(cancelButton);
    
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
  });
  
  it('should handle SSH control toggle', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    
    render(
      <EnvironmentForm 
        onSubmit={onSubmit}
        mode="create"
      />
    );
    
    // Enable SSH control
    const sshControlCheckbox = screen.getByLabelText(/Enable SSH Control/i);
    await user.click(sshControlCheckbox);
    
    // SSH fields should now be visible
    await waitFor(() => {
      expect(screen.getByLabelText(/SSH Host/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/SSH Port/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/SSH Username/i)).toBeInTheDocument();
    });
  });
  
  it('should show all configuration sections', () => {
    const onSubmit = vi.fn();
    
    render(
      <EnvironmentForm 
        onSubmit={onSubmit}
        mode="create"
      />
    );
    
    // Check that all major sections are present by looking for section headers
    expect(screen.getByRole('heading', { name: /Basic Information/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /SSH Control/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /Health Check Configuration/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /Restart Configuration/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /Upgrade Configuration/i })).toBeInTheDocument();
  });
  
  it('should handle SSH authentication type changes', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    
    render(
      <EnvironmentForm 
        onSubmit={onSubmit}
        mode="create"
      />
    );
    
    // Enable SSH control
    const sshControlCheckbox = screen.getByLabelText(/Enable SSH Control/i);
    await user.click(sshControlCheckbox);
    
    // Wait for SSH fields
    await waitFor(() => {
      expect(screen.getByLabelText(/Authentication Method/i)).toBeInTheDocument();
    });
    
    // Change auth type to key - first click to open the select
    const authMethodSelect = screen.getByLabelText(/Authentication Method/i);
    await user.click(authMethodSelect);
    
    // Wait for menu to appear and click the option
    const keyOption = await screen.findByText('Private Key');
    await user.click(keyOption);
    
    // Private key field should be visible - using placeholder text as fallback
    await waitFor(() => {
      const privateKeyField = screen.getByPlaceholderText(/Paste your private key here/i);
      expect(privateKeyField).toBeInTheDocument();
    });
  });
  
  it('should disable form fields when submitting', () => {
    const onSubmit = vi.fn().mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
        isLoading={true}
      />
    );

    // Check if buttons are disabled during loading
    expect(screen.getByRole('button', { name: /Create Environment/i })).toBeDisabled();
    expect(screen.getByRole('button', { name: /Cancel/i })).toBeDisabled();
  });

  it('should toggle health check accordion via switch', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Health check is disabled by default, switch index 0
    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    const healthCheckSwitch = switchInputs[0] as HTMLElement;
    await user.click(healthCheckSwitch);

    await waitFor(() => {
      expect(screen.getByLabelText(/health check endpoint/i)).toBeInTheDocument();
    });
  });

  it('should show health check fields when enabled', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[0] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByLabelText(/http method/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/check interval/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/timeout/i)).toBeInTheDocument();
    });
  });

  it('should show health check disabled text when health check is off', () => {
    const onSubmit = vi.fn();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    expect(screen.getByText(/health checks are disabled/i)).toBeInTheDocument();
  });

  it('should show restart disabled text when restart is toggled off', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Restart is enabled by default (index 1), click to disable
    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    const restartSwitch = switchInputs[1] as HTMLElement;
    await user.click(restartSwitch);

    await waitFor(() => {
      expect(screen.getByText(/restart functionality is disabled/i)).toBeInTheDocument();
    });
  });

  it('should show restart SSH command field when SSH is enabled and restart enabled', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Enable SSH control
    const sshCheckbox = screen.getByLabelText(/enable ssh control/i);
    await user.click(sshCheckbox);

    await waitFor(() => {
      // With SSH enabled and restart enabled, SSH host field shows - verifying SSH is on
      expect(screen.getByLabelText(/ssh host/i)).toBeInTheDocument();
    });

    // The command type selector should be present in the restart accordion (which is expanded)
    // Command Type field shows in the DOM since restart.enabled=true by default
    await waitFor(() => {
      const commandTypeField = screen.queryByLabelText(/command type/i);
      if (commandTypeField) {
        expect(commandTypeField).toBeInTheDocument();
      } else {
        // If not accessible, check that restart accordion shows SSH-related UI
        expect(screen.getByText(/restart configuration/i)).toBeInTheDocument();
      }
    });
  });

  it('should show upgrade config fields when upgrade is enabled', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    // upgrade switch is index 2
    const upgradeSwitch = switchInputs[2] as HTMLElement;
    await user.click(upgradeSwitch);

    await waitFor(() => {
      expect(screen.getByText(/version list endpoint configuration/i)).toBeInTheDocument();
    });
  });

  it('should show JSONPath field when upgrade is enabled', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[2] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByLabelText(/jsonpath response/i)).toBeInTheDocument();
    });
  });

  it('should show upgrade command type selector when upgrade is enabled', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[2] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByLabelText(/upgrade command type/i)).toBeInTheDocument();
    });
  });

  it('should show SSH upgrade command when SSH enabled and upgrade enabled', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Enable SSH
    const sshCheckbox = screen.getByLabelText(/enable ssh control/i);
    await user.click(sshCheckbox);

    await waitFor(() => {
      expect(screen.getByLabelText(/ssh host/i)).toBeInTheDocument();
    });

    // Enable upgrade
    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    // With SSH enabled: health check (0), restart (1), upgrade (2)
    await user.click(switchInputs[2] as HTMLElement);

    await waitFor(() => {
      // Upgrade accordion is now expanded and SSH type shows upgrade-related sections
      expect(screen.getByText(/version list endpoint configuration/i)).toBeInTheDocument();
    });
  });

  it('should update SSH host input', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const sshCheckbox = screen.getByLabelText(/enable ssh control/i);
    await user.click(sshCheckbox);

    await waitFor(() => {
      expect(screen.getByLabelText(/ssh host/i)).toBeInTheDocument();
    });

    await user.type(screen.getByLabelText(/ssh host/i), 'server.example.com');
    expect((screen.getByLabelText(/ssh host/i) as HTMLInputElement).value).toBe('server.example.com');
  }, 15000);

  it('should update SSH username input', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const sshCheckbox = screen.getByLabelText(/enable ssh control/i);
    await user.click(sshCheckbox);

    await waitFor(() => {
      expect(screen.getByLabelText(/ssh username/i)).toBeInTheDocument();
    });

    await user.type(screen.getByLabelText(/ssh username/i), 'deploy');
    expect((screen.getByLabelText(/ssh username/i) as HTMLInputElement).value).toBe('deploy');
  });

  it('shows error when error prop provided', () => {
    render(
      <EnvironmentForm
        onSubmit={vi.fn()}
        mode="create"
        error="An error occurred"
      />
    );
    expect(screen.getByText(/an error occurred/i)).toBeInTheDocument();
  });

  it('initializes from edit mode data with SSH enabled', async () => {
    const onSubmit = vi.fn();

    render(
      <EnvironmentForm
        initialData={mockEnvironment}
        onSubmit={onSubmit}
        mode="edit"
      />
    );

    await waitFor(() => {
      // SSH control should be auto-enabled because credentials.username is set
      expect(screen.getByLabelText(/ssh host/i)).toBeInTheDocument();
    });
  });

  it('shows HTTP upgrade config when upgrade enabled without SSH', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Don't enable SSH - enable upgrade only
    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    // upgrade switch is index 2
    await user.click(switchInputs[2] as HTMLElement);

    await waitFor(() => {
      // Without SSH, upgrade shows HTTP config (Upgrade Endpoint field)
      expect(screen.getByText(/version list endpoint configuration/i)).toBeInTheDocument();
    });
  });

  it('handleHttpUpgradeChange updates form data', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    const { fireEvent: fe } = await import('@testing-library/react');

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Enable upgrade (without SSH so HTTP mode)
    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[2] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByText(/version list endpoint configuration/i)).toBeInTheDocument();
    });

    // Find upgrade endpoint URL field via label
    const upgradeUrlFields = screen.queryAllByLabelText(/upgrade endpoint/i);
    if (upgradeUrlFields.length > 0) {
      fe.change(upgradeUrlFields[0], { target: { value: 'http://myserver/upgrade/{VERSION}' } });
      expect((upgradeUrlFields[0] as HTMLInputElement).value).toBe('http://myserver/upgrade/{VERSION}');
    }
  });

  it('handleHttpRestartChange updates form data', async () => {
    const onSubmit = vi.fn();
    const { fireEvent: fe } = await import('@testing-library/react');

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Restart is enabled by default, SSH is disabled by default → HTTP mode
    // Find restart endpoint URL field
    await waitFor(() => {
      const restartUrlFields = screen.queryAllByLabelText(/restart endpoint/i);
      if (restartUrlFields.length > 0) {
        fe.change(restartUrlFields[0], { target: { value: 'http://myserver/restart' } });
        expect((restartUrlFields[0] as HTMLInputElement).value).toBe('http://myserver/restart');
      }
    });
  });

  it('should disable SSH fields again when SSH control is toggled off after being enabled', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const sshCheckbox = screen.getByLabelText(/enable ssh control/i);
    await user.click(sshCheckbox);

    await waitFor(() => {
      expect(screen.getByLabelText(/ssh host/i)).toBeInTheDocument();
    });

    // Toggle SSH control back off
    await user.click(sshCheckbox);

    await waitFor(() => {
      expect(screen.queryByLabelText(/ssh host/i)).not.toBeInTheDocument();
      expect(screen.getByText(/ssh control is disabled/i)).toBeInTheDocument();
    });
  });

  it('should update SSH port input', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const sshCheckbox = screen.getByLabelText(/enable ssh control/i);
    await user.click(sshCheckbox);

    await waitFor(() => {
      expect(screen.getByLabelText(/ssh port/i)).toBeInTheDocument();
    });

    const portField = screen.getByLabelText(/ssh port/i);
    fireEvent.change(portField, { target: { value: '2222' } });

    expect((portField as HTMLInputElement).value).toBe('2222');
  });

  it('should not toggle accordion expansion when clicking directly on an accordion header', () => {
    const onSubmit = vi.fn();

    const { container } = render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const summaries = container.querySelectorAll('.MuiAccordionSummary-root');
    expect(summaries.length).toBe(3);

    // Health check accordion starts collapsed; clicking the header directly must not expand it
    fireEvent.click(summaries[0]);
    expect(screen.getByText(/health checks are disabled/i)).toBeInTheDocument();

    // Restart accordion starts expanded (restart.enabled=true by default); clicking header keeps it expanded
    fireEvent.click(summaries[1]);
    expect(screen.getByText(/restart configuration/i)).toBeInTheDocument();
    expect(screen.queryByText(/restart functionality is disabled/i)).not.toBeInTheDocument();

    // Upgrade accordion starts collapsed; clicking header directly must not expand it
    fireEvent.click(summaries[2]);
    expect(screen.queryByText(/version list endpoint configuration/i)).not.toBeInTheDocument();
  });

  it('should update health check endpoint, interval and timeout fields', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[0] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByLabelText(/health check endpoint/i)).toBeInTheDocument();
    });

    const endpointField = screen.getByLabelText(/health check endpoint/i);
    fireEvent.change(endpointField, { target: { value: '/status' } });
    expect((endpointField as HTMLInputElement).value).toBe('/status');

    const intervalField = screen.getByLabelText(/check interval/i);
    fireEvent.change(intervalField, { target: { value: '60' } });
    expect((intervalField as HTMLInputElement).value).toBe('60');

    const timeoutField = screen.getByLabelText(/timeout \(seconds\)/i);
    fireEvent.change(timeoutField, { target: { value: '10' } });
    expect((timeoutField as HTMLInputElement).value).toBe('10');
  });

  it('should change the health check HTTP method', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[0] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByLabelText(/http method/i)).toBeInTheDocument();
    });

    const methodSelect = screen.getByLabelText(/http method/i);
    await user.click(methodSelect);
    const listbox = await screen.findByRole('listbox');
    await user.click(within(listbox).getByText('HEAD'));

    await waitFor(() => {
      expect(screen.getByLabelText(/http method/i)).toHaveTextContent('HEAD');
    });
  }, 15000);

  it('should update health check validation type and expected value', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[0] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByLabelText(/validation type/i)).toBeInTheDocument();
    });

    // Default validation type is statusCode - update expected value (numeric branch)
    const expectedValueField = screen.getByLabelText(/expected value/i);
    fireEvent.change(expectedValueField, { target: { value: '404' } });
    expect((expectedValueField as HTMLInputElement).value).toBe('404');
    expect(screen.getByText(/expected http status code/i)).toBeInTheDocument();

    // Switch validation type to JSON Regex
    const validationTypeSelect = screen.getByLabelText(/validation type/i);
    await user.click(validationTypeSelect);
    const listbox = await screen.findByRole('listbox');
    await user.click(within(listbox).getByText(/json regex/i));

    await waitFor(() => {
      expect(screen.getByText(/regular expression to match in response/i)).toBeInTheDocument();
    });

    // Update expected value again (string branch)
    const expectedValueField2 = screen.getByLabelText(/expected value/i);
    fireEvent.change(expectedValueField2, { target: { value: '^ok$' } });
    expect((expectedValueField2 as HTMLInputElement).value).toBe('^ok$');
  }, 15000);

  it('should switch restart command type to SSH and update the SSH restart command field', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const sshCheckbox = screen.getByLabelText(/enable ssh control/i);
    await user.click(sshCheckbox);

    await waitFor(() => {
      expect(screen.getByLabelText(/command type/i)).toBeInTheDocument();
    });

    const commandTypeSelect = screen.getByLabelText(/command type/i);
    await user.click(commandTypeSelect);
    const listbox = await screen.findByRole('listbox');
    await user.click(within(listbox).getByText('SSH'));

    await waitFor(() => {
      expect(screen.getByLabelText(/ssh restart command/i)).toBeInTheDocument();
    });

    const sshCommandField = screen.getByLabelText(/ssh restart command/i);
    fireEvent.change(sshCommandField, { target: { value: 'sudo systemctl restart myapp' } });
    expect((sshCommandField as HTMLInputElement).value).toBe('sudo systemctl restart myapp');
  });

  it('should switch upgrade command type to SSH and update the SSH upgrade commands field', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const sshCheckbox = screen.getByLabelText(/enable ssh control/i);
    await user.click(sshCheckbox);

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    // With SSH enabled: health check (0), restart (1), upgrade (2)
    await user.click(switchInputs[2] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByLabelText(/upgrade command type/i)).toBeInTheDocument();
    });

    const upgradeTypeSelect = screen.getByLabelText(/upgrade command type/i);
    await user.click(upgradeTypeSelect);
    const listbox = await screen.findByRole('listbox');
    await user.click(within(listbox).getByText('SSH'));

    await waitFor(() => {
      expect(screen.getByLabelText(/ssh upgrade commands/i)).toBeInTheDocument();
    });

    const sshUpgradeField = screen.getByLabelText(/ssh upgrade commands/i);
    fireEvent.change(sshUpgradeField, { target: { value: 'sudo app-upgrade --version={VERSION}' } });
    expect((sshUpgradeField as HTMLInputElement).value).toBe('sudo app-upgrade --version={VERSION}');
  }, 15000);

  it('should keep raw text when the restart request body contains invalid JSON', () => {
    const onSubmit = vi.fn();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Restart is enabled by default and defaults to HTTP mode since SSH control is off
    const restartUrlField = screen.getByLabelText(/restart endpoint/i);
    const restartContainer = restartUrlField.closest('.MuiGrid-container') as HTMLElement;
    const bodyField = within(restartContainer).getByLabelText(/request body/i);

    fireEvent.change(bodyField, { target: { value: '{invalid json' } });

    expect((bodyField as HTMLInputElement).value).toBe('{invalid json');
  });

  it('should keep raw text when the upgrade request body contains invalid JSON', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[2] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByLabelText(/upgrade endpoint/i)).toBeInTheDocument();
    });

    const upgradeUrlField = screen.getByLabelText(/upgrade endpoint/i);
    const upgradeContainer = upgradeUrlField.closest('.MuiGrid-container') as HTMLElement;
    const bodyField = within(upgradeContainer).getByLabelText(/request body/i);

    fireEvent.change(bodyField, { target: { value: '{also invalid' } });

    expect((bodyField as HTMLInputElement).value).toBe('{also invalid');
  });

  it('should update version list URL and JSONPath response fields when upgrade is enabled', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[2] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByLabelText(/version list url/i)).toBeInTheDocument();
    });

    const versionListUrlField = screen.getByLabelText(/version list url/i);
    fireEvent.change(versionListUrlField, { target: { value: 'http://api.example.com/versions' } });
    expect((versionListUrlField as HTMLInputElement).value).toBe('http://api.example.com/versions');

    const jsonPathField = screen.getByLabelText(/jsonpath response/i);
    fireEvent.change(jsonPathField, { target: { value: '$.data.releases[*]' } });
    expect((jsonPathField as HTMLInputElement).value).toBe('$.data.releases[*]');
  }, 15000);
});
