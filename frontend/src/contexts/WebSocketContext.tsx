import React, { createContext, useContext, useEffect, useRef, useState } from 'react';
import { useAppDispatch } from '@/store';
import { updateEnvironmentStatus } from '@/store/slices/environmentSlice';
import { showInfo, showError } from '@/store/slices/notificationSlice';

interface WebSocketContextType {
  isConnected: boolean;
  subscribe: (environmentIds: string[]) => void;
  unsubscribe: (environmentIds: string[]) => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within WebSocketProvider');
  }
  return context;
};

interface WebSocketProviderProps {
  children: React.ReactNode;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({ children }) => {
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const dispatch = useAppDispatch();

  const connect = () => {
    try {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const ws = new WebSocket(`${protocol}//${window.location.host}/ws`);

      ws.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          handleMessage(message);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        dispatch(showError('Real-time connection error'));
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
        setIsConnected(false);
        wsRef.current = null;

        // Attempt to reconnect after 5 seconds
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, 5000);
      };

      wsRef.current = ws;
    } catch (error) {
      console.error('Failed to connect WebSocket:', error);
    }
  };

  const handleMessage = (message: any) => {
    switch (message.type) {
      case 'status_update':
        if (message.payload?.environmentId && message.payload?.status) {
          dispatch(updateEnvironmentStatus({
            id: message.payload.environmentId,
            status: message.payload.status,
          }));
        }
        break;

      case 'operation_update':
        // Handle operation updates
        const { operationId, update } = message.payload || {};
        if (update?.status === 'completed') {
          dispatch(showInfo(`Operation ${operationId} completed successfully`));
        } else if (update?.status === 'failed') {
          dispatch(showError(`Operation ${operationId} failed: ${update.error}`));
        }
        break;

      case 'pong':
        // Handle pong response
        break;

      default:
        console.warn('Unknown WebSocket message type:', message.type);
    }
  };

  const subscribe = (environmentIds: string[]) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'subscribe',
        payload: { environments: environmentIds },
      }));
    }
  };

  const unsubscribe = (environmentIds: string[]) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'unsubscribe',
        payload: { environments: environmentIds },
      }));
    }
  };

  useEffect(() => {
    connect();

    // Ping interval to keep connection alive
    const pingInterval = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'ping' }));
      }
    }, 30000);

    return () => {
      clearInterval(pingInterval);
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  return (
    <WebSocketContext.Provider value={{ isConnected, subscribe, unsubscribe }}>
      {children}
    </WebSocketContext.Provider>
  );
};
