import React from 'react';
import {
  Grid,
  TextField,
  MenuItem,
} from '@mui/material';

export interface HttpRequestData {
  url?: string;
  method?: string;
  headers?: Record<string, string>;
  body?: string;
}

interface HttpRequestConfigProps {
  data: HttpRequestData;
  onChange: (data: HttpRequestData) => void;
  urlLabel?: string;
  urlPlaceholder?: string;
  urlHelperText?: string;
  showBody?: boolean;
  required?: boolean;
}

const HTTP_METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'];

export const HttpRequestConfig: React.FC<HttpRequestConfigProps> = ({
  data,
  onChange,
  urlLabel = 'URL',
  urlPlaceholder = 'https://api.example.com/endpoint',
  urlHelperText,
  showBody = true,
  required = false,
}) => {
  // Local state for headers text to avoid escaping issues
  const [headersText, setHeadersText] = React.useState('');
  const [headersError, setHeadersError] = React.useState('');
  
  // Initialize headers text when data changes
  React.useEffect(() => {
    if (data.headers && typeof data.headers === 'object') {
      try {
        setHeadersText(JSON.stringify(data.headers, null, 2));
      } catch {
        setHeadersText('');
      }
    }
  }, [data.headers]);

  const handleChange = (field: keyof HttpRequestData, value: any) => {
    onChange({
      ...data,
      [field]: value,
    });
  };

  const handleHeadersChange = (value: string) => {
    setHeadersText(value);
    setHeadersError('');
    
    // Try to parse as JSON
    if (value.trim() === '') {
      handleChange('headers', {});
      return;
    }
    
    try {
      const parsed = JSON.parse(value);
      if (typeof parsed === 'object' && !Array.isArray(parsed)) {
        handleChange('headers', parsed);
      } else {
        setHeadersError('Headers must be a valid JSON object');
      }
    } catch (e) {
      setHeadersError('Invalid JSON format');
    }
  };

  return (
    <Grid container spacing={2}>
      <Grid item xs={12} md={8}>
        <TextField
          label={urlLabel}
          fullWidth
          required={required}
          value={data.url || ''}
          onChange={(e) => handleChange('url', e.target.value)}
          placeholder={urlPlaceholder}
          helperText={urlHelperText}
        />
      </Grid>
      
      <Grid item xs={12} md={4}>
        <TextField
          select
          label="Method"
          fullWidth
          required={required}
          value={data.method || 'GET'}
          onChange={(e) => handleChange('method', e.target.value)}
        >
          {HTTP_METHODS.map((method) => (
            <MenuItem key={method} value={method}>
              {method}
            </MenuItem>
          ))}
        </TextField>
      </Grid>
      
      <Grid item xs={12}>
        <TextField
          label="Headers (JSON)"
          fullWidth
          multiline
          rows={3}
          value={headersText}
          onChange={(e) => handleHeadersChange(e.target.value)}
          placeholder='{"Authorization": "Bearer token", "Content-Type": "application/json"}'
          helperText={headersError || "Optional: HTTP headers as JSON object"}
          error={!!headersError}
        />
      </Grid>
      
      {showBody && (
        <Grid item xs={12}>
          <TextField
            label="Request Body"
            fullWidth
            multiline
            rows={4}
            value={data.body || ''}
            onChange={(e) => handleChange('body', e.target.value)}
            placeholder='{"key": "value"}'
            helperText="Optional: Request body (for POST, PUT, PATCH methods)"
          />
        </Grid>
      )}
    </Grid>
  );
};
