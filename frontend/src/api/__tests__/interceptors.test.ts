/**
 * Tests for axios interceptors defined in environments.ts
 * These cover the request/response interceptor code (lines 66-91)
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('axios interceptors (environments.ts)', () => {
  let requestSuccessHandler: ((config: any) => any) | undefined;
  let requestErrorHandler: ((error: any) => any) | undefined;
  let responseSuccessHandler: ((response: any) => any) | undefined;
  let responseErrorHandler: ((error: any) => any) | undefined;

  beforeEach(() => {
    // Capture interceptor handlers by mocking axios.interceptors before importing
    vi.doMock('axios', () => {
      const interceptors = {
        request: {
          use: vi.fn((onSuccess: any, onError: any) => {
            requestSuccessHandler = onSuccess;
            requestErrorHandler = onError;
            return 0;
          }),
        },
        response: {
          use: vi.fn((onSuccess: any, onError: any) => {
            responseSuccessHandler = onSuccess;
            responseErrorHandler = onError;
            return 0;
          }),
        },
      };

      return {
        default: {
          get: vi.fn(),
          post: vi.fn(),
          put: vi.fn(),
          delete: vi.fn(),
          interceptors,
        },
        interceptors,
      };
    });
  });

  afterEach(() => {
    vi.doUnmock('axios');
    vi.resetModules();
    requestSuccessHandler = undefined;
    requestErrorHandler = undefined;
    responseSuccessHandler = undefined;
    responseErrorHandler = undefined;
  });

  it('request interceptor adds Authorization header when token exists', async () => {
    localStorage.setItem('authToken', 'test-token-123');

    await import('../environments');

    expect(requestSuccessHandler).toBeDefined();
    const config = { headers: {} as any };
    const result = requestSuccessHandler!(config);
    expect(result.headers.Authorization).toBe('Bearer test-token-123');

    localStorage.removeItem('authToken');
  });

  it('request interceptor does not add header when no token', async () => {
    localStorage.removeItem('authToken');

    await import('../environments');

    expect(requestSuccessHandler).toBeDefined();
    const config = { headers: {} as any };
    const result = requestSuccessHandler!(config);
    expect(result.headers.Authorization).toBeUndefined();
  });

  it('request error handler rejects with error', async () => {
    await import('../environments');

    expect(requestErrorHandler).toBeDefined();
    const error = new Error('request error');
    await expect(requestErrorHandler!(error)).rejects.toThrow('request error');
  });

  it('response success handler returns response unchanged', async () => {
    await import('../environments');

    expect(responseSuccessHandler).toBeDefined();
    const response = { data: { success: true }, status: 200 };
    const result = responseSuccessHandler!(response);
    expect(result).toBe(response);
  });

  it('response error handler handles 401 by removing token and redirecting', async () => {
    localStorage.setItem('authToken', 'my-token');
    const originalLocation = window.location.href;

    // Mock window.location
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { href: '' },
    });

    await import('../environments');

    expect(responseErrorHandler).toBeDefined();
    const error = { response: { status: 401 } };
    await expect(responseErrorHandler!(error)).rejects.toEqual(error);

    expect(localStorage.getItem('authToken')).toBeNull();
    expect(window.location.href).toBe('/login');

    // Restore
    window.location.href = originalLocation;
  });

  it('response error handler rejects non-401 errors without redirecting', async () => {
    localStorage.setItem('authToken', 'keep-this-token');

    Object.defineProperty(window, 'location', {
      writable: true,
      value: { href: '' },
    });

    await import('../environments');

    expect(responseErrorHandler).toBeDefined();
    const error = { response: { status: 500 } };
    await expect(responseErrorHandler!(error)).rejects.toEqual(error);

    // Token should still be there
    expect(localStorage.getItem('authToken')).toBe('keep-this-token');
    expect(window.location.href).not.toBe('/login');

    localStorage.removeItem('authToken');
  });

  it('response error handler handles error with no response object', async () => {
    await import('../environments');

    expect(responseErrorHandler).toBeDefined();
    const error = new Error('network error');
    await expect(responseErrorHandler!(error)).rejects.toEqual(error);
  });
});
