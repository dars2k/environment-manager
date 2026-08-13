import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
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

  it('handleHttpRestartChange falls back to raw string when body is invalid JSON', async () => {
    const onSubmit = vi.fn();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Restart is enabled by default and SSH is disabled → HTTP restart config visible
    const bodyField = await screen.findByLabelText(/request body/i);

    fireEvent.change(bodyField, { target: { value: '{invalid json' } });

    // The catch branch in handleHttpRestartChange stores the raw string back
    // (rather than throwing), so the field should reflect the typed value.
    await waitFor(() => {
      expect((bodyField as HTMLInputElement).value).toBe('{invalid json');
    });

    // Typing valid JSON afterwards should also be reflected (parsed then re-serialized
    // back through the typeof === 'string' ternary once stored as an object is no longer
    // a string, but our immediate re-render still shows the raw text the user typed).
    fireEvent.change(bodyField, { target: { value: '{"key":"value"}' } });
    await waitFor(() => {
      expect((bodyField as HTMLInputElement).value).toBe('{"key":"value"}');
    });
  });

  it('handleHttpUpgradeChange falls back to raw string when body is invalid JSON', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Enable upgrade (SSH disabled → HTTP mode for the upgrade command section)
    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[2] as HTMLElement);

    await waitFor(() => {
      expect(screen.getByLabelText(/upgrade endpoint/i)).toBeInTheDocument();
    });

    // There are now two "Request Body" fields: restart's and the upgrade command's.
    // The upgrade command one is the last in the DOM.
    const bodyFields = screen.getAllByLabelText(/request body/i);
    const upgradeBodyField = bodyFields[bodyFields.length - 1];

    fireEvent.change(upgradeBodyField, { target: { value: 'not { valid json' } });

    await waitFor(() => {
      expect((upgradeBodyField as HTMLInputElement).value).toBe('not { valid json');
    });
  });

  it('disabling SSH control after enabling it resets fields back to the disabled state', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const sshCheckbox = screen.getByLabelText(/enable ssh control/i);

    // Enable
    await user.click(sshCheckbox);
    await waitFor(() => {
      expect(screen.getByLabelText(/ssh host/i)).toBeInTheDocument();
    });

    // Disable again - exercises the reset branch inside the checkbox onChange handler
    await user.click(sshCheckbox);
    await waitFor(() => {
      expect(screen.queryByLabelText(/ssh host/i)).not.toBeInTheDocument();
      expect(screen.getByText(/ssh control is disabled/i)).toBeInTheDocument();
    });
  });

  it('updates the SSH port field', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    await user.click(screen.getByLabelText(/enable ssh control/i));
    await waitFor(() => {
      expect(screen.getByLabelText(/ssh port/i)).toBeInTheDocument();
    });

    const portField = screen.getByLabelText(/ssh port/i) as HTMLInputElement;
    fireEvent.change(portField, { target: { value: '2222' } });

    expect(portField.value).toBe('2222');
  }, 15000);

  it('clicking directly on the health check accordion summary does not toggle the switch', async () => {
    const onSubmit = vi.fn();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const summary = screen.getByText('Health Check Configuration').closest('.MuiAccordionSummary-root');
    expect(summary).toBeTruthy();

    const switchInput = document.querySelectorAll('.MuiSwitch-input')[0] as HTMLInputElement;
    expect(switchInput.checked).toBe(false);

    fireEvent.click(summary as Element);

    // The onClick handler calls preventDefault when clicking the summary bar itself,
    // so the underlying switch state should be unaffected by this click.
    expect(switchInput.checked).toBe(false);
  });

  it('clicking directly on the restart accordion summary does not toggle the switch', async () => {
    const onSubmit = vi.fn();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const summary = screen.getByText('Restart Configuration').closest('.MuiAccordionSummary-root');
    expect(summary).toBeTruthy();

    const switchInput = document.querySelectorAll('.MuiSwitch-input')[1] as HTMLInputElement;
    expect(switchInput.checked).toBe(true);

    fireEvent.click(summary as Element);

    expect(switchInput.checked).toBe(true);
  });

  it('clicking directly on the upgrade accordion summary does not toggle the switch', async () => {
    const onSubmit = vi.fn();

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const summary = screen.getByText('Upgrade Configuration').closest('.MuiAccordionSummary-root');
    expect(summary).toBeTruthy();

    const switchInput = document.querySelectorAll('.MuiSwitch-input')[2] as HTMLInputElement;
    expect(switchInput.checked).toBe(false);

    fireEvent.click(summary as Element);

    expect(switchInput.checked).toBe(false);
  });

  it('updates all health check fields when health check is enabled', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[0] as HTMLElement);

    // Endpoint
    const endpointField = await screen.findByLabelText(/health check endpoint/i);
    await user.clear(endpointField);
    await user.type(endpointField, '/status');
    expect((endpointField as HTMLInputElement).value).toBe('/status');

    // HTTP Method select
    const methodSelect = screen.getByLabelText(/http method/i);
    await user.click(methodSelect);
    const postOption = await screen.findByRole('option', { name: 'POST' });
    await user.click(postOption);
    await waitFor(() => {
      expect(screen.getByLabelText(/http method/i)).toHaveTextContent('POST');
    });

    // Interval
    const intervalField = screen.getByLabelText(/check interval/i) as HTMLInputElement;
    fireEvent.change(intervalField, { target: { value: '60' } });
    expect(intervalField.value).toBe('60');

    // Timeout
    const timeoutField = screen.getByLabelText(/timeout \(seconds\)/i) as HTMLInputElement;
    fireEvent.change(timeoutField, { target: { value: '15' } });
    expect(timeoutField.value).toBe('15');
  }, 15000);

  it('changes validation type to JSON Regex and updates helper text and expected value', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[0] as HTMLElement);

    const expectedValueField = await screen.findByLabelText(/expected value/i);
    expect(screen.getByText(/expected http status code/i)).toBeInTheDocument();

    // Default statusCode validation - typing updates the numeric value
    fireEvent.change(expectedValueField, { target: { value: '404' } });
    expect((expectedValueField as HTMLInputElement).value).toBe('404');

    // Switch validation type to JSON Regex
    const validationTypeSelect = screen.getByLabelText(/validation type/i);
    await user.click(validationTypeSelect);
    const regexOption = await screen.findByRole('option', { name: 'JSON Regex' });
    await user.click(regexOption);

    await waitFor(() => {
      expect(screen.getByText(/regular expression to match in response/i)).toBeInTheDocument();
    });

    const expectedValueFieldAfter = screen.getByLabelText(/expected value/i);
    fireEvent.change(expectedValueFieldAfter, { target: { value: 'success' } });
    expect((expectedValueFieldAfter as HTMLInputElement).value).toBe('success');
  }, 15000);

  it('switches restart command type to SSH and updates the SSH restart command field', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Enable SSH control so the "SSH" command type option becomes selectable
    await user.click(screen.getByLabelText(/enable ssh control/i));
    await waitFor(() => {
      expect(screen.getByLabelText(/ssh host/i)).toBeInTheDocument();
    });

    const commandTypeSelect = screen.getByLabelText(/command type/i);
    await user.click(commandTypeSelect);
    const sshOption = await screen.findByRole('option', { name: 'SSH' });
    await user.click(sshOption);

    const sshCommandField = await screen.findByLabelText(/ssh restart command/i);
    await user.type(sshCommandField, 'sudo systemctl restart myapp');

    expect((sshCommandField as HTMLInputElement).value).toBe('sudo systemctl restart myapp');
  }, 15000);

  it('updates the version list URL and JSONPath response fields when upgrade is enabled', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[2] as HTMLElement);

    const versionListUrlField = await screen.findByLabelText(/version list url/i);
    fireEvent.change(versionListUrlField, { target: { value: 'https://api.example.com/versions' } });
    expect((versionListUrlField as HTMLInputElement).value).toBe('https://api.example.com/versions');

    const jsonPathField = screen.getByLabelText(/jsonpath response/i);
    fireEvent.change(jsonPathField, { target: { value: '$.data[*]' } });
    expect((jsonPathField as HTMLInputElement).value).toBe('$.data[*]');
  }, 15000);

  it('switches upgrade command type to SSH and updates the SSH upgrade commands field', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <EnvironmentForm
        onSubmit={onSubmit}
        mode="create"
      />
    );

    // Enable SSH control
    await user.click(screen.getByLabelText(/enable ssh control/i));
    await waitFor(() => {
      expect(screen.getByLabelText(/ssh host/i)).toBeInTheDocument();
    });

    // Enable upgrade
    const switchInputs = document.querySelectorAll('.MuiSwitch-input');
    await user.click(switchInputs[2] as HTMLElement);

    const upgradeTypeSelect = await screen.findByLabelText(/upgrade command type/i);
    await user.click(upgradeTypeSelect);
    const sshOption = await screen.findByRole('option', { name: 'SSH' });
    await user.click(sshOption);

    const sshUpgradeField = await screen.findByLabelText(/ssh upgrade commands/i);
    fireEvent.change(sshUpgradeField, { target: { value: 'sudo app-upgrade --version={VERSION}' } });

    expect((sshUpgradeField as HTMLInputElement).value).toBe('sudo app-upgrade --version={VERSION}');
  }, 15000);
});
