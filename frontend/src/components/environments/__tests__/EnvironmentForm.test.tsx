import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@/test/test-utils';
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
    const user = userEvent.setup();
    
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
  });
  
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
});
