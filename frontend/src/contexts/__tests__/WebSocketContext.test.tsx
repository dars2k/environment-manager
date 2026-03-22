import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { WebSocketProvider, useWebSocket } from '../WebSocketContext';
import { createTestStore } from '@/test/test-utils';
// Mock WebSocket
const mockWsSend = vi.fn();
const mockWsClose = vi.fn();

class MockWebSocket {
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;

  readyState: number;
  onopen: (() => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: ((e: any) => void) | null = null;
  onmessage: ((e: MessageEvent) => void) | null = null;
  send = mockWsSend;
  close = mockWsClose;

  constructor(public url: string) {
    this.readyState = MockWebSocket.OPEN;
    setTimeout(() => {
      if (this.onopen) this.onopen();
    }, 0);
  }
}

vi.stubGlobal('WebSocket', MockWebSocket);

function TestConsumer() {
  const { isConnected, subscribe, unsubscribe } = useWebSocket();
  return (
    <div>
      <span data-testid="connected">{isConnected ? 'connected' : 'disconnected'}</span>
      <button onClick={() => subscribe(['env-1'])}>subscribe</button>
      <button onClick={() => unsubscribe(['env-1'])}>unsubscribe</button>
    </div>
  );
}

function renderWithStore() {
  const store = createTestStore();
  localStorage.setItem('authToken', 'test-token');
  return render(
    <Provider store={store}>
      <WebSocketProvider>
        <TestConsumer />
      </WebSocketProvider>
    </Provider>
  );
}

describe('WebSocketContext', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.setItem('authToken', 'test-token');
  });

  it('provides isConnected state', async () => {
    renderWithStore();
    await act(async () => {
      await new Promise(r => setTimeout(r, 50));
    });
    expect(screen.getByTestId('connected')).toBeInTheDocument();
  });

  it('throws when used outside provider', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    expect(() => {
      render(<TestConsumer />);
    }).toThrow('useWebSocket must be used within WebSocketProvider');
    spy.mockRestore();
  });

  it('calls send with subscribe message', async () => {
    renderWithStore();
    await act(async () => {
      await new Promise(r => setTimeout(r, 50));
    });
    const btn = screen.getAllByRole('button').find(b => b.textContent === 'subscribe')!;
    btn.click();
    expect(mockWsSend).toHaveBeenCalledWith(
      JSON.stringify({ type: 'subscribe', payload: { environments: ['env-1'] } })
    );
  });

  it('calls send with unsubscribe message', async () => {
    renderWithStore();
    await act(async () => {
      await new Promise(r => setTimeout(r, 50));
    });
    const btn = screen.getAllByRole('button').find(b => b.textContent === 'unsubscribe')!;
    btn.click();
    expect(mockWsSend).toHaveBeenCalledWith(
      JSON.stringify({ type: 'unsubscribe', payload: { environments: ['env-1'] } })
    );
  });

  it('skips connection when no auth token', async () => {
    localStorage.removeItem('authToken');
    const store = createTestStore();
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
    render(
      <Provider store={store}>
        <WebSocketProvider>
          <div data-testid="child">child</div>
        </WebSocketProvider>
      </Provider>
    );
    await act(async () => {
      await new Promise(r => setTimeout(r, 50));
    });
    expect(screen.getByTestId('child')).toBeInTheDocument();
    warnSpy.mockRestore();
  });

  it('dispatches updateEnvironmentStatus on status_update message', async () => {
    const store = createTestStore();
    let capturedWs: any;
    const OrigWS = MockWebSocket;
    const InterceptWS = class extends OrigWS {
      constructor(url: string) {
        super(url);
        capturedWs = this;
      }
    };
    vi.stubGlobal('WebSocket', InterceptWS);
    localStorage.setItem('authToken', 'test-token');

    render(
      <Provider store={store}>
        <WebSocketProvider>
          <div data-testid="child">child</div>
        </WebSocketProvider>
      </Provider>
    );

    await act(async () => {
      await new Promise(r => setTimeout(r, 50));
    });

    await act(async () => {
      capturedWs.onmessage({
        data: JSON.stringify({
          type: 'status_update',
          payload: { environmentId: 'env-1', status: { health: 'healthy' } },
        }),
      });
    });

    // Should not throw; store gets updated
    expect(screen.getByTestId('child')).toBeInTheDocument();
    vi.stubGlobal('WebSocket', MockWebSocket);
  });

  it('handles invalid JSON WebSocket message gracefully', async () => {
    const store = createTestStore();
    let capturedWs: any;
    const InterceptWS = class extends MockWebSocket {
      constructor(url: string) {
        super(url);
        capturedWs = this;
      }
    };
    vi.stubGlobal('WebSocket', InterceptWS);
    localStorage.setItem('authToken', 'test-token');
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <Provider store={store}>
        <WebSocketProvider>
          <div data-testid="child">child</div>
        </WebSocketProvider>
      </Provider>
    );

    await act(async () => {
      await new Promise(r => setTimeout(r, 50));
    });

    await act(async () => {
      capturedWs.onmessage({ data: 'not valid json {{{' });
    });

    expect(errorSpy).toHaveBeenCalled();
    errorSpy.mockRestore();
    vi.stubGlobal('WebSocket', MockWebSocket);
  });

  it('handles operation_update completed message', async () => {
    const store = createTestStore();
    let capturedWs: any;
    const InterceptWS = class extends MockWebSocket {
      constructor(url: string) { super(url); capturedWs = this; }
    };
    vi.stubGlobal('WebSocket', InterceptWS);
    localStorage.setItem('authToken', 'test-token');

    render(
      <Provider store={store}>
        <WebSocketProvider>
          <div data-testid="child">child</div>
        </WebSocketProvider>
      </Provider>
    );
    await act(async () => { await new Promise(r => setTimeout(r, 50)); });

    await act(async () => {
      capturedWs.onmessage({
        data: JSON.stringify({
          type: 'operation_update',
          payload: { operationId: 'op-1', update: { status: 'completed' } },
        }),
      });
    });

    // Should not throw
    expect(screen.getByTestId('child')).toBeInTheDocument();
    vi.stubGlobal('WebSocket', MockWebSocket);
  });

  it('handles operation_update failed message', async () => {
    const store = createTestStore();
    let capturedWs: any;
    const InterceptWS = class extends MockWebSocket {
      constructor(url: string) { super(url); capturedWs = this; }
    };
    vi.stubGlobal('WebSocket', InterceptWS);
    localStorage.setItem('authToken', 'test-token');

    render(
      <Provider store={store}>
        <WebSocketProvider>
          <div data-testid="child">child</div>
        </WebSocketProvider>
      </Provider>
    );
    await act(async () => { await new Promise(r => setTimeout(r, 50)); });

    await act(async () => {
      capturedWs.onmessage({
        data: JSON.stringify({
          type: 'operation_update',
          payload: { operationId: 'op-1', update: { status: 'failed', error: 'Timeout' } },
        }),
      });
    });

    expect(screen.getByTestId('child')).toBeInTheDocument();
    vi.stubGlobal('WebSocket', MockWebSocket);
  });

  it('handles pong message without error', async () => {
    const store = createTestStore();
    let capturedWs: any;
    const InterceptWS = class extends MockWebSocket {
      constructor(url: string) { super(url); capturedWs = this; }
    };
    vi.stubGlobal('WebSocket', InterceptWS);
    localStorage.setItem('authToken', 'test-token');

    render(
      <Provider store={store}>
        <WebSocketProvider>
          <div data-testid="child">child</div>
        </WebSocketProvider>
      </Provider>
    );
    await act(async () => { await new Promise(r => setTimeout(r, 50)); });

    await act(async () => {
      capturedWs.onmessage({ data: JSON.stringify({ type: 'pong' }) });
    });

    expect(screen.getByTestId('child')).toBeInTheDocument();
    vi.stubGlobal('WebSocket', MockWebSocket);
  });

  it('handles unknown message type with warning', async () => {
    const store = createTestStore();
    let capturedWs: any;
    const InterceptWS = class extends MockWebSocket {
      constructor(url: string) { super(url); capturedWs = this; }
    };
    vi.stubGlobal('WebSocket', InterceptWS);
    localStorage.setItem('authToken', 'test-token');
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

    render(
      <Provider store={store}>
        <WebSocketProvider>
          <div data-testid="child">child</div>
        </WebSocketProvider>
      </Provider>
    );
    await act(async () => { await new Promise(r => setTimeout(r, 50)); });

    await act(async () => {
      capturedWs.onmessage({ data: JSON.stringify({ type: 'unknown_type' }) });
    });

    expect(warnSpy).toHaveBeenCalled();
    warnSpy.mockRestore();
    vi.stubGlobal('WebSocket', MockWebSocket);
  });

  it('handles WebSocket error event', async () => {
    const store = createTestStore();
    let capturedWs: any;
    const InterceptWS = class extends MockWebSocket {
      constructor(url: string) { super(url); capturedWs = this; }
    };
    vi.stubGlobal('WebSocket', InterceptWS);
    localStorage.setItem('authToken', 'test-token');
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <Provider store={store}>
        <WebSocketProvider>
          <div data-testid="child">child</div>
        </WebSocketProvider>
      </Provider>
    );
    await act(async () => { await new Promise(r => setTimeout(r, 50)); });

    await act(async () => {
      capturedWs.onerror(new Error('Connection failed'));
    });

    expect(errorSpy).toHaveBeenCalled();
    errorSpy.mockRestore();
    vi.stubGlobal('WebSocket', MockWebSocket);
  });
});
