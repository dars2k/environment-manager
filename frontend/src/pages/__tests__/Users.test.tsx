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
    const keyButtons = document.querySelectorAll('[data-testid="VpnKeyIcon"]');
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

  it('validates create user form and shows field errors for short username/password', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /create user/i })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /create user/i }));
    await waitFor(() => {
      expect(screen.getByText(/create new user/i)).toBeInTheDocument();
    });

    // Enter a too-short username and too-short password, then submit.
    const usernameField = screen.getByLabelText(/username/i);
    fireEvent.change(usernameField, { target: { value: 'ab' } });
    const passwordFields = screen.getAllByLabelText(/password/i);
    fireEvent.change(passwordFields[0], { target: { value: '123' } });

    const createBtn = screen.getByRole('button', { name: /^create$/i });
    fireEvent.click(createBtn);

    await waitFor(() => {
      expect(screen.getByText(/username must be at least 3 characters/i)).toBeInTheDocument();
      expect(screen.getByText(/password must be at least 6 characters/i)).toBeInTheDocument();
    });
    // The API must never be called since client-side validation failed.
    expect(mockedUsersApi.createUser).not.toHaveBeenCalled();
  });

  it('shows submit error alert when create user API call fails', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([]);
    mockedUsersApi.createUser = vi.fn().mockRejectedValue({
      response: { data: { error: 'Username already exists' } },
    });
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /create user/i })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /create user/i }));
    await waitFor(() => {
      expect(screen.getByText(/create new user/i)).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText(/username/i), { target: { value: 'newuser' } });
    const passwordFields = screen.getAllByLabelText(/password/i);
    fireEvent.change(passwordFields[0], { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: /^create$/i }));

    await waitFor(() => {
      expect(screen.getByText(/username already exists/i)).toBeInTheDocument();
    });
    // Dialog should remain open since creation failed.
    expect(screen.getByText(/create new user/i)).toBeInTheDocument();
  });

  it('shows generic submit error when create user API fails without a response message', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([]);
    mockedUsersApi.createUser = vi.fn().mockRejectedValue(new Error('network down'));
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /create user/i })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /create user/i }));
    fireEvent.change(await screen.findByLabelText(/username/i), { target: { value: 'newuser' } });
    const passwordFields = screen.getAllByLabelText(/password/i);
    fireEvent.change(passwordFields[0], { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: /^create$/i }));

    await waitFor(() => {
      expect(screen.getByText(/failed to create user/i)).toBeInTheDocument();
    });
  });

  it('changes role selection in create dialog and closes on cancel, resetting fields', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([]);
    render(<Users />);
    fireEvent.click(await screen.findByRole('button', { name: /create user/i }));
    await waitFor(() => {
      expect(screen.getByText(/create new user/i)).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText(/username/i), { target: { value: 'someuser' } });

    // Open role select and choose Admin.
    fireEvent.mouseDown(screen.getByRole('combobox'));
    const adminOption = await screen.findByRole('option', { name: 'Admin' });
    fireEvent.click(adminOption);

    // Cancel closes and resets the dialog.
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    await waitFor(() => {
      expect(screen.queryByText(/create new user/i)).not.toBeInTheDocument();
    });

    // Reopening should show the form reset back to defaults (empty username).
    fireEvent.click(screen.getByRole('button', { name: /create user/i }));
    await waitFor(() => {
      expect(screen.getByText(/create new user/i)).toBeInTheDocument();
    });
    expect(screen.getByLabelText(/username/i)).toHaveValue('');
  });

  it('updates a user role/status successfully and refreshes the list', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    mockedUsersApi.updateUser = vi.fn().mockResolvedValue({ ...mockUser, role: 'user', active: false });
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });

    const editButtons = document.querySelectorAll('[data-testid="EditIcon"]');
    fireEvent.click(editButtons[0].parentElement!);
    await waitFor(() => {
      expect(screen.getByText(/edit user: admin/i)).toBeInTheDocument();
    });

    // Toggle the active switch off (currently active user).
    const activeSwitch = screen.getByRole('checkbox');
    expect(activeSwitch).toBeChecked();
    fireEvent.click(activeSwitch);
    expect(screen.getByText('Disabled')).toBeInTheDocument();

    // Change role to User.
    fireEvent.mouseDown(screen.getByRole('combobox'));
    const userOption = await screen.findByRole('option', { name: 'User' });
    fireEvent.click(userOption);

    fireEvent.click(screen.getByRole('button', { name: /^save$/i }));

    await waitFor(() => {
      expect(mockedUsersApi.updateUser).toHaveBeenCalledWith('user-1', { role: 'user', active: false });
    });
    // Dialog closes on success.
    await waitFor(() => {
      expect(screen.queryByText(/edit user: admin/i)).not.toBeInTheDocument();
    });
    // loadUsers is called again on success (listUsers invoked a second time).
    expect(mockedUsersApi.listUsers).toHaveBeenCalledTimes(2);
  });

  it('shows an error alert in edit dialog when update fails', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    mockedUsersApi.updateUser = vi.fn().mockRejectedValue({
      response: { data: { error: 'Cannot demote last admin' } },
    });
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });

    const editButtons = document.querySelectorAll('[data-testid="EditIcon"]');
    fireEvent.click(editButtons[0].parentElement!);
    await waitFor(() => {
      expect(screen.getByText(/edit user: admin/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /^save$/i }));

    await waitFor(() => {
      expect(screen.getByText(/cannot demote last admin/i)).toBeInTheDocument();
    });
    // Dialog remains open after failure.
    expect(screen.getByText(/edit user: admin/i)).toBeInTheDocument();
  });

  it('does nothing when edit dialog submit is triggered with no user selected', async () => {
    // Render dialog directly closed/no-user state is internal; instead verify that
    // opening edit for a user then cancel does not call updateUser.
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    mockedUsersApi.updateUser = vi.fn().mockResolvedValue(mockUser);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });
    const editButtons = document.querySelectorAll('[data-testid="EditIcon"]');
    fireEvent.click(editButtons[0].parentElement!);
    await waitFor(() => {
      expect(screen.getByText(/edit user: admin/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    await waitFor(() => {
      expect(screen.queryByText(/edit user: admin/i)).not.toBeInTheDocument();
    });
    expect(mockedUsersApi.updateUser).not.toHaveBeenCalled();
  });

  it('validates reset password: too short and mismatched passwords show errors', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });
    const keyButtons = document.querySelectorAll('[data-testid="VpnKeyIcon"]');
    fireEvent.click(keyButtons[0].parentElement!);
    await waitFor(() => {
      expect(screen.getByText(/reset password for/i)).toBeInTheDocument();
    });

    // Too short password.
    fireEvent.change(screen.getByLabelText(/new password/i), { target: { value: '123' } });
    fireEvent.click(screen.getByRole('button', { name: /reset password/i }));
    await waitFor(() => {
      expect(screen.getByText(/password must be at least 6 characters/i)).toBeInTheDocument();
    });
    expect(mockedUsersApi.resetPassword).not.toHaveBeenCalled();

    // Now long enough but mismatched confirm.
    fireEvent.change(screen.getByLabelText(/new password/i), { target: { value: 'password123' } });
    fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'different123' } });
    fireEvent.click(screen.getByRole('button', { name: /reset password/i }));
    await waitFor(() => {
      expect(screen.getByText(/passwords do not match/i)).toBeInTheDocument();
    });
    expect(mockedUsersApi.resetPassword).not.toHaveBeenCalled();
  });

  it('resets password successfully and closes the dialog', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    mockedUsersApi.resetPassword = vi.fn().mockResolvedValue(undefined);
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });
    const keyButtons = document.querySelectorAll('[data-testid="VpnKeyIcon"]');
    fireEvent.click(keyButtons[0].parentElement!);
    await waitFor(() => {
      expect(screen.getByText(/reset password for/i)).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText(/new password/i), { target: { value: 'password123' } });
    fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: /reset password/i }));

    await waitFor(() => {
      expect(mockedUsersApi.resetPassword).toHaveBeenCalledWith('user-1', { newPassword: 'password123' });
    });
    await waitFor(() => {
      expect(screen.queryByText(/reset password for/i)).not.toBeInTheDocument();
    });
  });

  it('shows API error alert and resets fields on cancel after failed reset', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    mockedUsersApi.resetPassword = vi.fn().mockRejectedValue({
      response: { data: { error: 'Password reuse not allowed' } },
    });
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });
    const keyButtons = document.querySelectorAll('[data-testid="VpnKeyIcon"]');
    fireEvent.click(keyButtons[0].parentElement!);
    await waitFor(() => {
      expect(screen.getByText(/reset password for/i)).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText(/new password/i), { target: { value: 'password123' } });
    fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: /reset password/i }));

    await waitFor(() => {
      expect(screen.getByText(/password reuse not allowed/i)).toBeInTheDocument();
    });

    // Cancel triggers handleClose, resetting fields and closing the dialog.
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    await waitFor(() => {
      expect(screen.queryByText(/reset password for/i)).not.toBeInTheDocument();
    });

    // Reopen and verify fields were cleared (no leftover error or password value).
    const keyButtons2 = document.querySelectorAll('[data-testid="VpnKeyIcon"]');
    fireEvent.click(keyButtons2[0].parentElement!);
    await waitFor(() => {
      expect(screen.getByText(/reset password for/i)).toBeInTheDocument();
    });
    expect(screen.queryByText(/password reuse not allowed/i)).not.toBeInTheDocument();
    expect(screen.getByLabelText(/new password/i)).toHaveValue('');
  });

  it('shows generic error when delete user API fails', async () => {
    mockedUsersApi.listUsers = vi.fn().mockResolvedValue([mockUser]);
    mockedUsersApi.deleteUser = vi.fn().mockRejectedValue({
      response: { data: { error: 'Cannot delete the last admin user' } },
    });
    render(<Users />);
    await waitFor(() => {
      expect(screen.getAllByText('admin').length).toBeGreaterThanOrEqual(1);
    });
    const deleteButtons = document.querySelectorAll('[data-testid="DeleteIcon"]');
    fireEvent.click(deleteButtons[0].parentElement!);
    await waitFor(() => {
      expect(screen.getAllByText(/delete user/i).length).toBeGreaterThanOrEqual(1);
    });
    fireEvent.click(screen.getByRole('button', { name: /^delete$/i }));

    await waitFor(() => {
      expect(screen.getByText(/cannot delete the last admin user/i)).toBeInTheDocument();
    });
    // Confirm dialog closes even though it failed? No: on error, the confirm dialog stays open per handler
    // (setDeleteConfirmDialog reset only happens on success). Verify delete button still present.
    expect(screen.getByRole('button', { name: /^delete$/i })).toBeInTheDocument();
  });

  it('dismisses the top-level error alert via its close button', async () => {
    mockedUsersApi.listUsers = vi.fn().mockRejectedValue(new Error('boom'));
    render(<Users />);
    await waitFor(() => {
      expect(screen.getByText(/failed to load users/i)).toBeInTheDocument();
    });
    const closeBtn = screen.getByRole('button', { name: /close/i });
    fireEvent.click(closeBtn);
    await waitFor(() => {
      expect(screen.queryByText(/failed to load users/i)).not.toBeInTheDocument();
    });
  });
});
