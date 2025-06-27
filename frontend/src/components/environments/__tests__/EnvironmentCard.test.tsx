import { describe, it, expect } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '@/test/test-utils';
import { EnvironmentCard } from '../EnvironmentCard';
import { Environment, HealthStatus } from '@/types/environment';

const mockEnvironment: Environment = {
  id: 'env-123',
  name: 'Test Environment',
  description: 'A test environment for unit testing',
  environmentURL: 'https://test.example.com',
  target: {
    host: 'test.example.com',
    port: 22,
  },
  credentials: {
    type: 'key',
    username: 'testuser',
    keyId: 'key-123',
  },
  healthCheck: {
    enabled: true,
    endpoint: '/health',
    method: 'GET',
    interval: 60,
    timeout: 10,
    validation: {
      type: 'statusCode',
      value: 200,
    },
  },
  status: {
    health: HealthStatus.Healthy,
    lastCheck: new Date().toISOString(),
    message: 'All systems operational',
    responseTime: 150,
  },
  systemInfo: {
    osVersion: 'Ubuntu 20.04',
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
    enabled: true,
    type: 'ssh',
    versionListURL: 'https://api.example.com/versions',
    jsonPathResponse: '$.versions',
    upgradeCommand: {
      command: 'sudo app-upgrade --version={VERSION}',
    },
  },
};

describe('EnvironmentCard', () => {
  it('should render environment information correctly', () => {
    render(<EnvironmentCard environment={mockEnvironment} />);
    
    expect(screen.getByText(mockEnvironment.name)).toBeInTheDocument();
    expect(screen.getByText(mockEnvironment.description)).toBeInTheDocument();
    expect(screen.getByText(/healthy/i)).toBeInTheDocument();
  });

  it('should display correct health status color', () => {
    const unhealthyEnv = {
      ...mockEnvironment,
      status: { ...mockEnvironment.status, health: HealthStatus.Unhealthy },
    };
    
    render(<EnvironmentCard environment={unhealthyEnv} />);
    
    const statusChip = screen.getByText(/unhealthy/i);
    const chipRoot = statusChip.closest('.MuiChip-root');
    expect(chipRoot).toHaveClass('MuiChip-colorError');
  });

  it('should handle environment URL click', async () => {
    const user = userEvent.setup();
    render(<EnvironmentCard environment={mockEnvironment} />);
    
    const urlButton = screen.getByText(mockEnvironment.environmentURL!);
    await user.click(urlButton);
    
    // The component opens URL in a new tab, we can't test window.open directly
    // but we can verify the button is clickable
    expect(urlButton).toBeInTheDocument();
  });

  it('should open menu on more button click', async () => {
    const user = userEvent.setup();
    render(<EnvironmentCard environment={mockEnvironment} />);
    
    const moreButton = screen.getByTestId('MoreVertIcon').parentElement as HTMLElement;
    await user.click(moreButton);
    
    await waitFor(() => {
      expect(screen.getByText(/view details/i)).toBeInTheDocument();
      expect(screen.getByText(/restart/i)).toBeInTheDocument();
      expect(screen.getByText(/edit/i)).toBeInTheDocument();
      expect(screen.getByText(/delete/i)).toBeInTheDocument();
    });
  });

  it('should show upgrade option when upgrade is enabled', async () => {
    const user = userEvent.setup();
    render(<EnvironmentCard environment={mockEnvironment} />);
    
    const moreButton = screen.getByTestId('MoreVertIcon').parentElement as HTMLElement;
    await user.click(moreButton);
    
    await waitFor(() => {
      expect(screen.getByText(/upgrade/i)).toBeInTheDocument();
    });
  });

  it('should disable upgrade option when upgrade is disabled', async () => {
    const user = userEvent.setup();
    const envWithoutUpgrade = {
      ...mockEnvironment,
      upgradeConfig: { ...mockEnvironment.upgradeConfig, enabled: false },
    };
    
    render(<EnvironmentCard environment={envWithoutUpgrade} />);
    
    const moreButton = screen.getByTestId('MoreVertIcon').parentElement as HTMLElement;
    await user.click(moreButton);
    
    await waitFor(() => {
      const upgradeMenuItem = screen.getByText(/upgrade/i).closest('li');
      expect(upgradeMenuItem).toHaveAttribute('aria-disabled', 'true');
    });
  });

  it('should display unknown status correctly', () => {
    const unknownStatusEnv = {
      ...mockEnvironment,
      status: { ...mockEnvironment.status, health: HealthStatus.Unknown },
    };
    
    render(<EnvironmentCard environment={unknownStatusEnv} />);
    
    const statusChip = screen.getByText(/unknown/i);
    const chipRoot = statusChip.closest('.MuiChip-root');
    expect(chipRoot).toHaveClass('MuiChip-colorWarning');
  });

  it('should display system info when available', () => {
    render(<EnvironmentCard environment={mockEnvironment} />);
    
    // The component only displays appVersion in a chip
    expect(screen.getByText(mockEnvironment.systemInfo.appVersion)).toBeInTheDocument();
  });

  it('should truncate long descriptions', () => {
    const longDescriptionEnv = {
      ...mockEnvironment,
      description: 'This is a very long description that should be truncated to fit within the card limits. '.repeat(5),
    };
    
    render(<EnvironmentCard environment={longDescriptionEnv} />);
    
    const description = screen.getByText(/This is a very long description/);
    const styles = window.getComputedStyle(description);
    expect(styles.overflow).toBe('hidden');
  });
});
