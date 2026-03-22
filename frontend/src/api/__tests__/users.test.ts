import { describe, it, expect, vi, beforeEach } from 'vitest';
import axios from 'axios';
import { usersApi } from '../users';

vi.mock('axios');
const mockedAxios = vi.mocked(axios);

const mockUser = {
  id: 'user-1',
  username: 'testuser',
  role: 'admin' as const,
  active: true,
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
};

describe('usersApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('listUsers returns array of users', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({ data: { data: { users: [mockUser] } } });

    const result = await usersApi.listUsers();
    expect(result).toHaveLength(1);
    expect(result[0].username).toBe('testuser');
    expect(mockedAxios.get).toHaveBeenCalledWith('/api/v1/users?page=1&limit=50');
  });

  it('listUsers with custom pagination params', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({ data: { data: { users: [] } } });

    await usersApi.listUsers(2, 25);
    expect(mockedAxios.get).toHaveBeenCalledWith('/api/v1/users?page=2&limit=25');
  });

  it('getUser returns single user', async () => {
    mockedAxios.get = vi.fn().mockResolvedValue({ data: { data: { user: mockUser } } });

    const result = await usersApi.getUser('user-1');
    expect(result.id).toBe('user-1');
  });

  it('createUser posts and returns user', async () => {
    mockedAxios.post = vi.fn().mockResolvedValue({ data: { data: { user: mockUser } } });

    const result = await usersApi.createUser({ username: 'testuser', password: 'pass123' });
    expect(result.username).toBe('testuser');
    expect(mockedAxios.post).toHaveBeenCalledWith('/api/v1/users', { username: 'testuser', password: 'pass123' });
  });

  it('createUser with role', async () => {
    mockedAxios.post = vi.fn().mockResolvedValue({ data: { data: { user: { ...mockUser, role: 'viewer' } } } });

    const result = await usersApi.createUser({ username: 'viewer1', password: 'pass', role: 'viewer' });
    expect(result.role).toBe('viewer');
  });

  it('updateUser puts and returns updated user', async () => {
    const updated = { ...mockUser, role: 'user' as const };
    mockedAxios.put = vi.fn().mockResolvedValue({ data: { data: { user: updated } } });

    const result = await usersApi.updateUser('user-1', { role: 'user' });
    expect(result.role).toBe('user');
    expect(mockedAxios.put).toHaveBeenCalledWith('/api/v1/users/user-1', { role: 'user' });
  });

  it('deleteUser calls delete endpoint', async () => {
    mockedAxios.delete = vi.fn().mockResolvedValue({});

    await expect(usersApi.deleteUser('user-1')).resolves.toBeUndefined();
    expect(mockedAxios.delete).toHaveBeenCalledWith('/api/v1/users/user-1');
  });

  it('resetPassword posts to password reset endpoint', async () => {
    mockedAxios.post = vi.fn().mockResolvedValue({});

    await expect(usersApi.resetPassword('user-1', { newPassword: 'newpass123' })).resolves.toBeUndefined();
    expect(mockedAxios.post).toHaveBeenCalledWith(
      '/api/v1/users/user-1/password/reset',
      { newPassword: 'newpass123' }
    );
  });
});
