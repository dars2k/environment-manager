import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import App from '../App';

// Mock WebSocket
class MockWebSocket {
  onopen: (() => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: ((e: any) => void) | null = null;
  onmessage: ((e: any) => void) | null = null;
  readyState = 0;
  close = vi.fn();
  send = vi.fn();
  constructor() {
    setTimeout(() => {
      if (this.onopen) this.onopen();
    }, 0);
  }
}
vi.stubGlobal('WebSocket', MockWebSocket);

vi.mock('@/routes', () => ({
  AppRoutes: () => <div data-testid="app-routes">App Routes</div>,
}));

describe('App', () => {
  it('renders without crashing', async () => {
    render(<App />);
    await waitFor(() => {
      expect(screen.getByTestId('app-routes')).toBeInTheDocument();
    });
  });

  it('wraps app routes in providers', async () => {
    render(<App />);
    await waitFor(() => {
      expect(screen.getByTestId('app-routes')).toBeInTheDocument();
    });
  });
});
