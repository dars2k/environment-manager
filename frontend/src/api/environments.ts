import axios from 'axios';
import { Environment, CreateEnvironmentRequest, UpdateEnvironmentRequest, OperationResponse, VersionsResponse, UpgradeRequest } from '@/types/environment';

const API_BASE_URL = '/api/v1';

interface ListEnvironmentsResponse {
  environments: Environment[];
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
}

export const environmentApi = {
  list: async (): Promise<ListEnvironmentsResponse> => {
    const response = await axios.get(`${API_BASE_URL}/environments`);
    const data = response.data.data;
    // Handle null environments by converting to empty array
    return {
      ...data,
      environments: data.environments || []
    };
  },

  get: async (id: string): Promise<Environment> => {
    const response = await axios.get(`${API_BASE_URL}/environments/${id}`);
    return response.data.data.environment;
  },

  create: async (data: CreateEnvironmentRequest): Promise<Environment> => {
    const response = await axios.post(`${API_BASE_URL}/environments`, data);
    return response.data.data;
  },

  update: async (id: string, data: UpdateEnvironmentRequest): Promise<Environment> => {
    const response = await axios.put(`${API_BASE_URL}/environments/${id}`, data);
    return response.data.data;
  },

  delete: async (id: string): Promise<void> => {
    await axios.delete(`${API_BASE_URL}/environments/${id}`);
  },

  restart: async (id: string, force: boolean = false): Promise<OperationResponse> => {
    const response = await axios.post(`${API_BASE_URL}/environments/${id}/restart`, { force });
    return response.data.data;
  },

  checkHealth: async (id: string): Promise<void> => {
    await axios.post(`${API_BASE_URL}/environments/${id}/check-health`);
  },

  getVersions: async (id: string): Promise<VersionsResponse> => {
    const response = await axios.get(`${API_BASE_URL}/environments/${id}/versions`);
    return response.data.data;
  },

  upgrade: async (id: string, data: UpgradeRequest): Promise<OperationResponse> => {
    const response = await axios.post(`${API_BASE_URL}/environments/${id}/upgrade`, data);
    return response.data.data;
  },
};

// Request interceptor for auth
axios.interceptors.request.use(
  (config) => {
    // Add auth token if available
    const token = localStorage.getItem('authToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized access
      localStorage.removeItem('authToken');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
