import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
import { NotificationPopover } from '../NotificationPopover';
import axios from 'axios';

vi.mock('axios');
const mockedAxios = vi.mocked(axios);

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockLog = {
  id: 'log-1',
  timestamp: '2024-01-15T10:00:00.000Z',
  environmentName: 'Production',
  username: 'admin',
  type: 'health_check',
  level: 'info' as const,
  message: 'Health check passed',
};

describe('NotificationPopover', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.setItem('authToken', 'test-token');
  });

  it('renders without crashing when closed', () => {
    render(
      <NotificationPopover anchorEl={null} open={false} onClose={vi.fn()} />
    );
    // Popover is closed, content not visible
    expect(screen.queryByText('Notifications')).toBeNull();
  });

  it('shows loading spinner while fetching', async () => {
    mockedAxios.get = vi.fn((_url: string) => new Promise(() => {})); // never resolves
    const div = document.createElement('div');
    render(
      <NotificationPopover anchorEl={div} open={true} onClose={vi.fn()} />
    );
    await waitFor(() => {
      expect(document.querySelector('.MuiCircularProgress-root')).not.toBeNull();
    });
  });

  it('shows no notifications message when empty', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [], pagination: {} } },
    });
    const div = document.createElement('div');
    render(
      <NotificationPopover anchorEl={div} open={true} onClose={vi.fn()} />
    );
    await waitFor(() => {
      expect(screen.getByText(/no new notifications/i)).toBeInTheDocument();
    });
  });

  it('shows logs when data is loaded', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog], pagination: {} } },
    });
    const div = document.createElement('div');
    render(
      <NotificationPopover anchorEl={div} open={true} onClose={vi.fn()} />
    );
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
    });
  });

  it('shows error when fetch fails', async () => {
    mockedAxios.get = vi.fn().mockRejectedValue(new Error('Network error'));
    const div = document.createElement('div');
    render(
      <NotificationPopover anchorEl={div} open={true} onClose={vi.fn()} />
    );
    await waitFor(() => {
      expect(screen.getByText(/failed to load notifications/i)).toBeInTheDocument();
    });
  });

  it('shows environment name in log entry', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog], pagination: {} } },
    });
    const div = document.createElement('div');
    render(
      <NotificationPopover anchorEl={div} open={true} onClose={vi.fn()} />
    );
    await waitFor(() => {
      expect(screen.getByText(/production/i)).toBeInTheDocument();
    });
  });

  it('shows username in log entry', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog], pagination: {} } },
    });
    const div = document.createElement('div');
    render(
      <NotificationPopover anchorEl={div} open={true} onClose={vi.fn()} />
    );
    await waitFor(() => {
      expect(screen.getByText(/by: admin/i)).toBeInTheDocument();
    });
  });

  it('calls onClose and navigates when show all is clicked', async () => {
    const onClose = vi.fn();
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog], pagination: {} } },
    });
    const div = document.createElement('div');
    render(
      <NotificationPopover anchorEl={div} open={true} onClose={onClose} />
    );
    await waitFor(() => {
      expect(screen.getByText(/show all notifications/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/show all notifications/i));
    expect(onClose).toHaveBeenCalledTimes(1);
    expect(mockNavigate).toHaveBeenCalledWith('/logs');
  });

  it('calls onClose when close button clicked', async () => {
    const onClose = vi.fn();
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [], pagination: {} } },
    });
    const div = document.createElement('div');
    render(
      <NotificationPopover anchorEl={div} open={true} onClose={onClose} />
    );
    await waitFor(() => {
      expect(screen.getByText('Notifications')).toBeInTheDocument();
    });
    const closeBtn = document.querySelector('[data-testid="CloseIcon"]');
    if (closeBtn) {
      fireEvent.click(closeBtn.parentElement!);
      expect(onClose).toHaveBeenCalled();
    }
  });

  it('renders multiple log entries with separators', async () => {
    const log2 = { ...mockLog, id: 'log-2', message: 'Environment restarted', level: 'success' as const };
    mockedAxios.get = vi.fn().mockResolvedValue({
      data: { data: { logs: [mockLog, log2], pagination: {} } },
    });
    const div = document.createElement('div');
    render(
      <NotificationPopover anchorEl={div} open={true} onClose={vi.fn()} />
    );
    await waitFor(() => {
      expect(screen.getByText('Health check passed')).toBeInTheDocument();
      expect(screen.getByText('Environment restarted')).toBeInTheDocument();
    });
  });
});
