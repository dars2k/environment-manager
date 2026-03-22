import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
import { EnvironmentLogs } from '../EnvironmentLogs';
import axios from 'axios';

vi.mock('axios');
const mockedAxios = vi.mocked(axios);

const mockLogs = [
  {
    id: 'log-1',
    timestamp: '2024-01-15T10:00:00.000Z',
    type: 'health_check',
    level: 'success',
    message: 'Health check passed',
    action: 'check',
  },
  {
    id: 'log-2',
    timestamp: '2024-01-15T10:05:00.000Z',
    type: 'action',
    level: 'error',
    message: 'Restart failed',
    details: { code: 500, reason: 'timeout' },
  },
  {
    id: 'log-3',
    timestamp: '2024-01-15T10:10:00.000Z',
    type: 'system',
    level: 'warning',
    message: 'High CPU usage',
  },
  {
    id: 'log-4',
    timestamp: '2024-01-15T10:15:00.000Z',
    type: 'auth',
    level: 'info',
    message: 'User logged in',
  },
];

describe('EnvironmentLogs', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows loading spinner while fetching', () => {
    mockedAxios.get = vi.fn((_url: string) => new Promise(() => {}));
    render(<EnvironmentLogs environmentId="env-1" />);
    expect(document.querySelector('.MuiCircularProgress-root')).not.toBeNull();
  });

  it('shows error when fetch fails', async () => {
    mockedAxios.get = vi.fn().mockRejectedValue(new Error('Network error'));
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      expect(screen.getByText(/failed to load logs/i)).toBeInTheDocument();
    });
  });

  it('shows no logs message when empty', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [] } },
    });
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      expect(screen.getByText(/no logs available/i)).toBeInTheDocument();
    });
  });

  it('renders log entries when loaded', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: mockLogs } },
    });
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
      expect(screen.getByText('Restart failed')).toBeInTheDocument();
    });
  });

  it('shows filter chips', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: mockLogs } },
    });
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      // Filter chips exist — there may be multiple elements with "error" (chip + level chip in row)
      expect(screen.getAllByText(/^error$/i).length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText(/^warning$/i).length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText(/^info$/i).length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText(/^success$/i).length).toBeGreaterThanOrEqual(1);
    });
  });

  it('filters logs by level when chip is clicked', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: mockLogs } },
    });
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
    });
    // Click the first 'error' chip (the filter chip in the toolbar)
    const errorChips = screen.getAllByText(/^error$/i);
    fireEvent.click(errorChips[0]);
    await waitFor(() => {
      expect(screen.getByText('Restart failed')).toBeInTheDocument();
      expect(screen.queryByText('Health check passed')).toBeNull();
    });
  });

  it('filters logs by search text', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: mockLogs } },
    });
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
    });
    const searchInput = screen.getByPlaceholderText(/search messages/i);
    fireEvent.change(searchInput, { target: { value: 'CPU' } });
    await waitFor(() => {
      expect(screen.getByText('High CPU usage')).toBeInTheDocument();
      expect(screen.queryByText('Health check passed')).toBeNull();
    });
  });

  it('shows no match message when filter yields no results', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: mockLogs } },
    });
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
    });
    const searchInput = screen.getByPlaceholderText(/search messages/i);
    fireEvent.change(searchInput, { target: { value: 'ZZZNOMATCH' } });
    await waitFor(() => {
      expect(screen.getByText(/no logs match the current filters/i)).toBeInTheDocument();
    });
  });

  it('shows column headers', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: mockLogs } },
    });
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      expect(screen.getByText('TIME')).toBeInTheDocument();
      expect(screen.getByText('TYPE')).toBeInTheDocument();
      expect(screen.getByText('LEVEL')).toBeInTheDocument();
      expect(screen.getByText('MESSAGE')).toBeInTheDocument();
    });
  });

  it('shows footer count', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: mockLogs } },
    });
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      expect(screen.getByText(/showing 4 of 4 log entries/i)).toBeInTheDocument();
    });
  });

  it('expands log details on click when details exist', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLogs[1]] } }, // log with details
    });
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      expect(screen.getByText('Restart failed')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText('Restart failed'));
    await waitFor(() => {
      expect(screen.getByText(/timeout/i)).toBeInTheDocument();
    });
  });

  it('resets to all filter on all chip click', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: mockLogs } },
    });
    render(<EnvironmentLogs environmentId="env-1" />);
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
    });
    // Filter to error first
    const errorChips = screen.getAllByText(/^error$/i);
    fireEvent.click(errorChips[0]);
    // Then back to all
    fireEvent.click(screen.getByText(/^all \(4\)/i));
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
      expect(screen.getByText('Restart failed')).toBeInTheDocument();
    });
  });
});
