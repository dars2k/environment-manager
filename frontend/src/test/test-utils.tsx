import React, { ReactElement } from 'react';
import { render as rtlRender, RenderOptions } from '@testing-library/react';
import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import { configureStore } from '@reduxjs/toolkit';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { WebSocketProvider } from '@/contexts/WebSocketContext';
import { darkTheme } from '@/theme';
import environmentReducer from '@/store/slices/environmentSlice';
import logsReducer from '@/store/slices/logsSlice';
import notificationReducer from '@/store/slices/notificationSlice';
import uiReducer from '@/store/slices/uiSlice';

// Create a custom store for testing
function createTestStore() {
  return configureStore({
    reducer: {
      environments: environmentReducer,
      logs: logsReducer,
      notifications: notificationReducer,
      ui: uiReducer,
    },
  });
}

export type TestStore = ReturnType<typeof createTestStore>;

interface AllTheProvidersProps {
  children: React.ReactNode;
  store?: TestStore;
}

const AllTheProviders: React.FC<AllTheProvidersProps> = ({ children, store }) => {
  const testStore = store || createTestStore();
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return (
    <Provider store={testStore}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter 
          future={{
            v7_startTransition: true,
            v7_relativeSplatPath: true,
          }}
        >
          <ThemeProvider theme={darkTheme}>
            <WebSocketProvider>{children}</WebSocketProvider>
          </ThemeProvider>
        </BrowserRouter>
      </QueryClientProvider>
    </Provider>
  );
};

const render = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'> & { store?: TestStore }
) => {
  const { store, ...renderOptions } = options || {};
  return rtlRender(ui, {
    wrapper: ({ children }) => <AllTheProviders store={store}>{children}</AllTheProviders>,
    ...renderOptions,
  });
};

// Re-export specific items from React Testing Library
export {
  screen,
  fireEvent,
  waitFor,
  waitForElementToBeRemoved,
  queryByAttribute,
  getByLabelText,
  getByText,
  getByTestId,
  getByDisplayValue,
  getByRole,
  getByTitle,
  getByPlaceholderText,
  getByAltText,
  getAllByLabelText,
  getAllByText,
  getAllByTitle,
  getAllByDisplayValue,
  getAllByRole,
  getAllByTestId,
  getAllByPlaceholderText,
  getAllByAltText,
  queryByLabelText,
  queryByText,
  queryByTestId,
  queryByDisplayValue,
  queryByRole,
  queryByTitle,
  queryByPlaceholderText,
  queryByAltText,
  queryAllByLabelText,
  queryAllByText,
  queryAllByTitle,
  queryAllByDisplayValue,
  queryAllByRole,
  queryAllByTestId,
  queryAllByPlaceholderText,
  queryAllByAltText,
  findByLabelText,
  findByText,
  findByTestId,
  findByDisplayValue,
  findByRole,
  findByTitle,
  findByPlaceholderText,
  findByAltText,
  findAllByLabelText,
  findAllByText,
  findAllByTitle,
  findAllByDisplayValue,
  findAllByRole,
  findAllByTestId,
  findAllByPlaceholderText,
  findAllByAltText,
  act,
  cleanup,
  within,
  prettyDOM,
} from '@testing-library/react';

// Export our custom render and createTestStore
export { render, createTestStore };
