import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@/test/test-utils';
import { Login } from '../Login';
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

describe('Login page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('renders login form elements', () => {
    render(<Login />);
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
    expect(screen.getByText(/Env Manager/i)).toBeInTheDocument();
  });

  it('sign in button is disabled when fields are empty', () => {
    render(<Login />);
    const button = screen.getByRole('button', { name: /sign in/i });
    expect(button).toBeDisabled();
  });

  it('sign in button enables when both fields have values', () => {
    render(<Login />);
    fireEvent.change(screen.getByLabelText(/username/i), { target: { value: 'admin' } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'pass' } });
    expect(screen.getByRole('button', { name: /sign in/i })).not.toBeDisabled();
  });

  it('updates username field on input', () => {
    render(<Login />);
    const usernameInput = screen.getByLabelText(/username/i);
    fireEvent.change(usernameInput, { target: { value: 'admin' } });
    expect(usernameInput).toHaveValue('admin');
  });

  it('updates password field on input', () => {
    render(<Login />);
    const passwordInput = screen.getByLabelText(/password/i);
    fireEvent.change(passwordInput, { target: { value: 'secret' } });
    expect(passwordInput).toHaveValue('secret');
  });

  it('toggles password visibility on icon button click', () => {
    render(<Login />);
    const passwordInput = screen.getByLabelText(/password/i);
    expect(passwordInput).toHaveAttribute('type', 'password');

    // Find the icon button inside the password adornment
    const toggleButton = screen.getByRole('button', { name: '' }) as HTMLButtonElement;
    // Actually find the visibility toggle - it's the only icon button besides submit
    const iconButtons = screen.getAllByRole('button');
    // The last non-submit button is the visibility toggle
    const visibilityButton = iconButtons.find(btn => btn !== screen.getByRole('button', { name: /sign in/i }));
    if (visibilityButton) {
      fireEvent.click(visibilityButton);
      expect(passwordInput).toHaveAttribute('type', 'text');
    }
  });

  it('shows error alert on failed login', async () => {
    mockedAxios.post = vi.fn().mockRejectedValue({
      response: { data: { error: { message: 'Invalid credentials' } } }
    });

    render(<Login />);
    fireEvent.change(screen.getByLabelText(/username/i), { target: { value: 'admin' } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'wrong' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(screen.getByText('Invalid credentials')).toBeInTheDocument();
    });
  });

  it('shows generic error when no specific message', async () => {
    mockedAxios.post = vi.fn().mockRejectedValue(new Error('Network error'));

    render(<Login />);
    fireEvent.change(screen.getByLabelText(/username/i), { target: { value: 'admin' } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'pass' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(screen.getByText(/login failed/i)).toBeInTheDocument();
    });
  });

  it('navigates to dashboard on successful login', async () => {
    mockedAxios.post = vi.fn().mockResolvedValue({
      data: { data: { token: 'jwt-token-123', user: { id: '1', username: 'admin' } } }
    });

    render(<Login />);
    fireEvent.change(screen.getByLabelText(/username/i), { target: { value: 'admin' } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'admin123' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });
    expect(localStorage.getItem('authToken')).toBe('jwt-token-123');
  });

  it('stores user in localStorage on successful login', async () => {
    const mockUser = { id: '1', username: 'admin', role: 'admin' };
    mockedAxios.post = vi.fn().mockResolvedValue({
      data: { data: { token: 'tok', user: mockUser } }
    });

    render(<Login />);
    fireEvent.change(screen.getByLabelText(/username/i), { target: { value: 'admin' } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'pass' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(localStorage.getItem('user')).toBe(JSON.stringify(mockUser));
    });
  });
});
