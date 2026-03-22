import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@/test/test-utils';
import { RestartEnvironmentDialog } from '../RestartEnvironmentDialog';
import { Environment, HealthStatus } from '@/types/environment';

const mockEnvironment: Environment = {
  id: 'env-1',
  name: 'Production',
  description: 'Prod',
  environmentURL: 'https://prod.example.com',
  target: { host: 'prod.example.com', port: 22 },
  credentials: { type: 'password', username: 'deploy' },
  healthCheck: { enabled: true, endpoint: '/health', method: 'GET', interval: 60, timeout: 10, validation: { type: 'statusCode', value: 200 } },
  status: { health: HealthStatus.Healthy, lastCheck: new Date().toISOString(), message: 'OK', responseTime: 100 },
  systemInfo: { osVersion: 'Ubuntu 22.04', appVersion: '3.1.0', lastUpdated: new Date().toISOString() },
  timestamps: { createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  commands: {
    type: 'ssh',
    restart: { enabled: true, command: 'systemctl restart app' },
  },
  upgradeConfig: { enabled: false, type: 'ssh', versionListURL: '', jsonPathResponse: '', upgradeCommand: {} },
};

describe('RestartEnvironmentDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('does not render content when closed', () => {
    render(
      <RestartEnvironmentDialog
        open={false}
        onClose={vi.fn()}
        environment={mockEnvironment}
        onRestart={vi.fn()}
      />
    );
    expect(screen.queryByText('Restart Environment')).toBeNull();
  });

  it('renders dialog title when open', () => {
    render(
      <RestartEnvironmentDialog
        open={true}
        onClose={vi.fn()}
        environment={mockEnvironment}
        onRestart={vi.fn()}
      />
    );
    expect(screen.getByText('Restart Environment')).toBeInTheDocument();
  });

  it('shows environment name', () => {
    render(
      <RestartEnvironmentDialog
        open={true}
        onClose={vi.fn()}
        environment={mockEnvironment}
        onRestart={vi.fn()}
      />
    );
    expect(screen.getAllByText('Production').length).toBeGreaterThanOrEqual(1);
  });

  it('shows warning when restart is disabled', () => {
    const disabledEnv = {
      ...mockEnvironment,
      commands: { ...mockEnvironment.commands, restart: { enabled: false } },
    };
    render(
      <RestartEnvironmentDialog
        open={true}
        onClose={vi.fn()}
        environment={disabledEnv}
        onRestart={vi.fn()}
      />
    );
    expect(screen.getByText(/restart disabled/i)).toBeInTheDocument();
  });

  it('shows Continue button when restart is enabled', () => {
    render(
      <RestartEnvironmentDialog
        open={true}
        onClose={vi.fn()}
        environment={mockEnvironment}
        onRestart={vi.fn()}
      />
    );
    expect(screen.getByRole('button', { name: /continue/i })).toBeInTheDocument();
  });

  it('proceeds to confirmation step on Continue click', () => {
    render(
      <RestartEnvironmentDialog
        open={true}
        onClose={vi.fn()}
        environment={mockEnvironment}
        onRestart={vi.fn()}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: /continue/i }));
    // "Confirm Restart" appears as both a heading and button
    expect(screen.getAllByText(/confirm restart/i).length).toBeGreaterThanOrEqual(1);
  });

  it('calls onRestart and onClose on confirm', () => {
    const onRestart = vi.fn();
    const onClose = vi.fn();
    render(
      <RestartEnvironmentDialog
        open={true}
        onClose={onClose}
        environment={mockEnvironment}
        onRestart={onRestart}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: /continue/i }));
    fireEvent.click(screen.getByRole('button', { name: /confirm restart/i }));
    expect(onRestart).toHaveBeenCalledWith(false); // forceRestart defaults to false
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('can go back from confirmation step', () => {
    render(
      <RestartEnvironmentDialog
        open={true}
        onClose={vi.fn()}
        environment={mockEnvironment}
        onRestart={vi.fn()}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: /continue/i }));
    expect(screen.getByRole('button', { name: /back/i })).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /back/i }));
    expect(screen.getByRole('button', { name: /continue/i })).toBeInTheDocument();
  });

  it('shows command preview for SSH', () => {
    render(
      <RestartEnvironmentDialog
        open={true}
        onClose={vi.fn()}
        environment={mockEnvironment}
        onRestart={vi.fn()}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: /continue/i }));
    expect(screen.getByText('systemctl restart app')).toBeInTheDocument();
  });

  it('calls onClose when Cancel button clicked', () => {
    const onClose = vi.fn();
    render(
      <RestartEnvironmentDialog
        open={true}
        onClose={onClose}
        environment={mockEnvironment}
        onRestart={vi.fn()}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('calls onRestart with force=true when force restart checked', () => {
    const onRestart = vi.fn();
    render(
      <RestartEnvironmentDialog
        open={true}
        onClose={vi.fn()}
        environment={mockEnvironment}
        onRestart={onRestart}
      />
    );
    // Check force restart checkbox
    const checkbox = screen.getByRole('checkbox');
    fireEvent.click(checkbox);
    fireEvent.click(screen.getByRole('button', { name: /continue/i }));
    fireEvent.click(screen.getByRole('button', { name: /confirm restart/i }));
    expect(onRestart).toHaveBeenCalledWith(true);
  });
});
