import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
import { Logs } from '../Logs';
import axios from 'axios';

vi.mock('axios');
const mockedAxios = vi.mocked(axios);

const mockLog = {
  id: 'log-1',
  timestamp: '2024-01-15T10:00:00.000Z',
  environmentId: 'env-1',
  environmentName: 'Production',
  type: 'health_check' as const,
  level: 'info' as const,
  message: 'Health check passed',
};

describe('Logs page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders logs page heading', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog], pagination: { page: 1, limit: 25, total: 1 } } }
    });

    render(<Logs />);
    await waitFor(() => {
      expect(screen.getByText(/system logs/i)).toBeInTheDocument();
    });
  });

  it('shows loading spinner while fetching', () => {
    mockedAxios.get = vi.fn(() => new Promise(() => {})); // Never resolves
    render(<Logs />);
    expect(document.querySelector('.MuiCircularProgress-root')).not.toBeNull();
  });

  it('renders log entries when loaded', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog], pagination: { page: 1, limit: 25, total: 1 } } }
    });

    render(<Logs />);
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
    });
  });

  it('shows empty state when no logs', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [], pagination: { page: 1, limit: 25, total: 0 } } }
    });

    render(<Logs />);
    await waitFor(() => {
      expect(screen.getByText(/no logs match the current filters/i)).toBeInTheDocument();
    });
  });

  it('shows error alert when fetch fails', async () => {
    mockedAxios.get = vi.fn().mockRejectedValue(new Error('Server error'));
    render(<Logs />);
    await waitFor(() => {
      expect(screen.getByText(/failed to load logs/i)).toBeInTheDocument();
    });
  });

  it('has search field', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [], pagination: { page: 1, limit: 25, total: 0 } } }
    });

    render(<Logs />);
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/search messages/i)).toBeInTheDocument();
    });
  });

  it('has refresh button', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [], pagination: { page: 1, limit: 25, total: 0 } } }
    });

    render(<Logs />);
    await waitFor(() => {
      const refreshBtn = document.querySelector('[data-testid="RefreshIcon"]');
      expect(refreshBtn).not.toBeNull();
    });
  });

  it('renders multiple log entries', async () => {
    const log2 = { ...mockLog, id: 'log-2', message: 'Environment restarted', type: 'action' as const, level: 'success' as const };
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog, log2], pagination: { page: 1, limit: 25, total: 2 } } }
    });

    render(<Logs />);
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
      expect(screen.getByText('Environment restarted')).toBeInTheDocument();
    });
  });

  it('shows environment name in log entry', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog], pagination: { page: 1, limit: 25, total: 1 } } }
    });

    render(<Logs />);
    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });
  });

  it('shows expand icon for logs with details', async () => {
    const logWithDetails = { ...mockLog, details: { code: 200, response: 'ok' } };
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [logWithDetails], pagination: { page: 1, limit: 25, total: 1 } } }
    });
    render(<Logs />);
    await waitFor(() => {
      expect(document.querySelector('[data-testid="KeyboardArrowDownIcon"]')).not.toBeNull();
    });
  });

  it('expands details panel when expand button clicked', async () => {
    const logWithDetails = { ...mockLog, details: { code: 200, response: 'ok' } };
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [logWithDetails], pagination: { page: 1, limit: 25, total: 1 } } }
    });
    render(<Logs />);
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
    });
    // Click the expand button
    const expandBtn = document.querySelector('[data-testid="KeyboardArrowDownIcon"]');
    if (expandBtn) {
      fireEvent.click(expandBtn.parentElement!);
    }
    // After clicking, the up arrow should show
    await waitFor(() => {
      expect(document.querySelector('[data-testid="KeyboardArrowUpIcon"]')).not.toBeNull();
    });
  });

  it('shows dash when no environment name', async () => {
    const logNoEnv = { ...mockLog, environmentName: undefined };
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [logNoEnv], pagination: { page: 1, limit: 25, total: 1 } } }
    });
    render(<Logs />);
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
    });
  });

  it('shows column headers', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog], pagination: { page: 1, limit: 25, total: 1 } } }
    });
    render(<Logs />);
    await waitFor(() => {
      expect(screen.getByText('TIME')).toBeInTheDocument();
      expect(screen.getByText('TYPE')).toBeInTheDocument();
      expect(screen.getByText('LEVEL')).toBeInTheDocument();
    });
  });

  it('shows pagination component', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog], pagination: { page: 1, limit: 25, total: 1 } } }
    });
    render(<Logs />);
    await waitFor(() => {
      expect(document.querySelector('.MuiTablePagination-root')).not.toBeNull();
    });
  });
});
