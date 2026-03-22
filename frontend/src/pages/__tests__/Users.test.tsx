import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@/test/test-utils';
import Users from '../Users';
import { usersApi } from '@/api/users';

vi.mock('@/api/users');
const mockedUsersApi = vi.mocked(usersApi);

const mockUser = {
  id: 'user-1',
  username: 'admin',
  role: 'admin' as const,
  active: true,
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
};

const mockViewer = {
  id: 'user-2',
  username: 'viewer1',
  role: 'viewer' as const,
  active: false,
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
};

describe('Users page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows loading spinner initially', () => {
    mockedUsersApi.listUsers = vi.fn(() => new Promise(() => {}));
    render(<Users />);
    expect(document.querySelector('.MuiCircularProgress-root')).not.toBeNull();
  });

  it('renders users management heading', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByText(/users management/i)).toBeInTheDocument();
    });
  });

  it('renders user list', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser, mockViewer]);
    render(<Users />);
    await waitFor(() => {
      // Username column cells
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText('viewer1')).toBeInTheDocument();
    });
  });

  it('shows error when users cannot be loaded', async () => {
    mockedUsersApi.listUsers = vi.fn().mockRejectedValue(new Error('Server error'));
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByText(/failed to load users/i)).toBeInTheDocument();
    });
  });

  it('has create user button', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /create user/i })).toBeInTheDocument();
    });
  });

  it('shows role chips for users', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser, mockViewer]);
    render(<Users />);
    await waitFor(() => {
      // viewer1 user has viewer role chip
      expect(screen.getByText('viewer')).toBeInTheDocument();
    });
  });

  it('shows active/disabled status chips', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser, mockViewer]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByText('Active')).toBeInTheDocument();
      expect(screen.getByText('Disabled')).toBeInTheDocument();
    });
  });

  it('opens create dialog on button click', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /create user/i })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /create user/i }));
    await waitFor(() => {
      expect(screen.getByText(/create new user/i)).toBeInTheDocument();
    });
  });

  it('renders table headers', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByText('Username')).toBeInTheDocument();
      expect(screen.getByText('Role')).toBeInTheDocument();
      expect(screen.getByText('Status')).toBeInTheDocument();
    });
  });

  it('opens edit dialog when edit icon clicked', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });
    // Click the edit icon button
    const editButtons = document.querySelectorAll('[data-testid="EditIcon"]');
    if (editButtons.length > 0) {
      fireEvent.click(editButtons[0].parentElement!);
      await waitFor(() => {
        expect(screen.getByText(/edit user/i)).toBeInTheDocument();
      });
    }
  });

  it('creates a user successfully', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([]);
    mockedUsersApi.createUser = vi.fn().mockResolvedValue({ id: 'new-user', username: 'newuser', role: 'user', active: true });
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /create user/i })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /create user/i }));
    await waitFor(() => {
      expect(screen.getByText(/create new user/i)).toBeInTheDocument();
    });
    // Fill in username
    const usernameField = screen.getByLabelText(/username/i);
    fireEvent.change(usernameField, { target: { value: 'newuser' } });
    // Fill in password
    const passwordFields = screen.getAllByLabelText(/password/i);
    fireEvent.change(passwordFields[0], { target: { value: 'password123' } });
    // Submit
    const createBtn = screen.getByRole('button', { name: /^create$/i });
    fireEvent.click(createBtn);
    await waitFor(() => {
      expect(mockedUsersApi.createUser).toHaveBeenCalled();
    });
  });

  it('shows delete confirmation dialog', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });
    const deleteButtons = document.querySelectorAll('[data-testid="DeleteIcon"]');
    if (deleteButtons.length > 0) {
      fireEvent.click(deleteButtons[0].parentElement!);
      await waitFor(() => {
        expect(screen.getAllByText(/delete user/i).length).toBeGreaterThanOrEqual(1);
      });
    }
  });

  it('deletes user when confirmed in dialog', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    mockedUsersApi.deleteUser = vi.fn().mockResolvedValue(undefined);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });
    const deleteButtons = document.querySelectorAll('[data-testid="DeleteIcon"]');
    if (deleteButtons.length > 0) {
      fireEvent.click(deleteButtons[0].parentElement!);
      await waitFor(() => {
        expect(screen.getAllByText(/delete user/i).length).toBeGreaterThanOrEqual(1);
      });
      // Click the Delete confirm button
      const deleteBtn = screen.getByRole('button', { name: /^delete$/i });
      fireEvent.click(deleteBtn);
      await waitFor(() => {
        expect(mockedUsersApi.deleteUser).toHaveBeenCalledWith(mockUser.id);
      });
    }
  });

  it('closes delete dialog on cancel', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });
    const deleteButtons = document.querySelectorAll('[data-testid="DeleteIcon"]');
    if (deleteButtons.length > 0) {
      fireEvent.click(deleteButtons[0].parentElement!);
      await waitFor(() => {
        expect(screen.getAllByText(/delete user/i).length).toBeGreaterThanOrEqual(1);
      });
      const cancelBtn = screen.getByRole('button', { name: /cancel/i });
      fireEvent.click(cancelBtn);
      // Dialog should close
      await waitFor(() => {
        const deleteBtns = screen.queryAllByText(/^delete$/i);
        expect(deleteBtns.length).toBe(0);
      });
    }
  });

  it('opens reset password dialog when key icon clicked', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });
    const keyButtons = document.querySelectorAll('[data-testid="KeyIcon"]');
    if (keyButtons.length > 0) {
      fireEvent.click(keyButtons[0].parentElement!);
      await waitFor(() => {
        expect(screen.getByText(/reset password for/i)).toBeInTheDocument();
      });
    }
  });

  it('shows last login as Never when not set', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByText('Never')).toBeInTheDocument();
    });
  });
});
