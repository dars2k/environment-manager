import axios from 'axios';

const API_BASE_URL = '/api/v1';

// User role types
export type UserRole = 'admin' | 'user' | 'viewer';

// User types
export interface User {
  id: string;
  username: string;
  role: UserRole;
  active: boolean;
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
  metadata?: Record<string, any>;
}

export interface CreateUserRequest {
  username: string;
  password: string;
}

export interface UpdateUserRequest {
  role?: UserRole;
  active?: boolean;
  metadata?: Record<string, any>;
}

export interface ResetPasswordRequest {
  newPassword: string;
}

export interface UsersResponse {
  data: {
    users: User[];
  };
}

export interface UserResponse {
  data: {
    user: User;
  };
}

// API functions
export const usersApi = {
  // List all users
  async listUsers(page = 1, limit = 50): Promise<User[]> {
    const response = await axios.get<UsersResponse>(
      `${API_BASE_URL}/users?page=${page}&limit=${limit}`
    );
    return response.data.data.users;
  },

  // Get a specific user
  async getUser(userId: string): Promise<User> {
    const response = await axios.get<UserResponse>(
      `${API_BASE_URL}/users/${userId}`
    );
    return response.data.data.user;
  },

  // Create a new user
  async createUser(request: CreateUserRequest): Promise<User> {
    const response = await axios.post<UserResponse>(
      `${API_BASE_URL}/users`,
      request
    );
    return response.data.data.user;
  },

  // Update a user
  async updateUser(userId: string, request: UpdateUserRequest): Promise<User> {
    const response = await axios.put<UserResponse>(
      `${API_BASE_URL}/users/${userId}`,
      request
    );
    return response.data.data.user;
  },

  // Delete a user
  async deleteUser(userId: string): Promise<void> {
    await axios.delete(`${API_BASE_URL}/users/${userId}`);
  },

  // Reset a user's password (admin only)
  async resetPassword(userId: string, request: ResetPasswordRequest): Promise<void> {
    await axios.post(
      `${API_BASE_URL}/users/${userId}/password/reset`,
      request
    );
  },
};
