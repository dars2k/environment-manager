import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@/test/test-utils';
import { HttpRequestConfig } from '../HttpRequestConfig';

const defaultData = {
  url: 'https://api.example.com/restart',
  method: 'POST',
  headers: { Authorization: 'Bearer token' },
  body: '{"action": "restart"}',
};

describe('HttpRequestConfig', () => {
  it('renders URL field', () => {
    render(<HttpRequestConfig data={defaultData} onChange={vi.fn()} />);
    expect(screen.getByLabelText(/^url$/i)).toHaveValue('https://api.example.com/restart');
  });

  it('renders method selector', () => {
    render(<HttpRequestConfig data={defaultData} onChange={vi.fn()} />);
    expect(screen.getByLabelText(/^method$/i)).toBeInTheDocument();
  });

  it('renders headers field with initial value', () => {
    render(<HttpRequestConfig data={defaultData} onChange={vi.fn()} />);
    const headersField = screen.getByLabelText(/headers/i);
    expect(headersField).toBeInTheDocument();
  });

  it('renders body field when showBody is true (default)', () => {
    render(<HttpRequestConfig data={defaultData} onChange={vi.fn()} />);
    expect(screen.getByLabelText(/request body/i)).toBeInTheDocument();
  });

  it('hides body field when showBody is false', () => {
    render(<HttpRequestConfig data={defaultData} onChange={vi.fn()} showBody={false} />);
    expect(screen.queryByLabelText(/request body/i)).toBeNull();
  });

  it('calls onChange when URL is changed', () => {
    const onChange = vi.fn();
    render(<HttpRequestConfig data={defaultData} onChange={onChange} />);
    fireEvent.change(screen.getByLabelText(/^url$/i), {
      target: { value: 'https://new-api.example.com' },
    });
    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({ url: 'https://new-api.example.com' })
    );
  });

  it('calls onChange when body is changed', () => {
    const onChange = vi.fn();
    render(<HttpRequestConfig data={defaultData} onChange={onChange} />);
    fireEvent.change(screen.getByLabelText(/request body/i), {
      target: { value: '{"action": "stop"}' },
    });
    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({ body: '{"action": "stop"}' })
    );
  });

  it('updates headers when valid JSON is entered', () => {
    const onChange = vi.fn();
    render(<HttpRequestConfig data={{ url: '', method: 'GET' }} onChange={onChange} />);
    const headersField = screen.getByLabelText(/headers/i);
    fireEvent.change(headersField, {
      target: { value: '{"Authorization": "Bearer abc"}' },
    });
    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({ headers: { Authorization: 'Bearer abc' } })
    );
  });

  it('shows error when invalid JSON entered for headers', () => {
    const onChange = vi.fn();
    render(<HttpRequestConfig data={{ url: '', method: 'GET' }} onChange={onChange} />);
    const headersField = screen.getByLabelText(/headers/i);
    fireEvent.change(headersField, {
      target: { value: 'not valid json' },
    });
    expect(screen.getByText(/invalid json format/i)).toBeInTheDocument();
  });

  it('clears headers when empty string entered', () => {
    const onChange = vi.fn();
    render(<HttpRequestConfig data={{ url: '', method: 'GET', headers: { key: 'val' } }} onChange={onChange} />);
    const headersField = screen.getByLabelText(/headers/i);
    fireEvent.change(headersField, { target: { value: '' } });
    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({ headers: {} })
    );
  });

  it('shows error when JSON array entered for headers', () => {
    const onChange = vi.fn();
    render(<HttpRequestConfig data={{ url: '', method: 'GET' }} onChange={onChange} />);
    const headersField = screen.getByLabelText(/headers/i);
    fireEvent.change(headersField, {
      target: { value: '[1, 2, 3]' },
    });
    expect(screen.getByText(/must be a valid json object/i)).toBeInTheDocument();
  });

  it('renders with custom URL label', () => {
    render(
      <HttpRequestConfig
        data={defaultData}
        onChange={vi.fn()}
        urlLabel="Endpoint URL"
      />
    );
    expect(screen.getByLabelText(/endpoint url/i)).toBeInTheDocument();
  });

  it('renders with required fields', () => {
    render(
      <HttpRequestConfig
        data={defaultData}
        onChange={vi.fn()}
        required={true}
      />
    );
    // Required fields have asterisks
    expect(screen.getByLabelText(/^url\s*\*/i)).toBeInTheDocument();
  });

  it('initializes headers text from data.headers prop', () => {
    render(
      <HttpRequestConfig
        data={{ url: '', method: 'GET', headers: { 'Content-Type': 'application/json' } }}
        onChange={vi.fn()}
      />
    );
    const headersField = screen.getByLabelText(/headers/i);
    expect((headersField as HTMLTextAreaElement).value).toContain('Content-Type');
  });
});
