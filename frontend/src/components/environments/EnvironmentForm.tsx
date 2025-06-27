import React, { useState, useEffect } from 'react';
import {
  Box,
  TextField,
  FormControlLabel,
  Switch,
  Checkbox,
  MenuItem,
  Grid,
  Typography,
  Paper,
  Divider,
  Alert,
  Button,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import { ExpandMore, Save, Cancel } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

import { Environment, CreateEnvironmentRequest } from '@/types/environment';
import { HttpRequestConfig, HttpRequestData } from './HttpRequestConfig';

interface EnvironmentFormProps {
  initialData?: Environment;
  onSubmit: (data: CreateEnvironmentRequest, password?: string, privateKey?: string) => Promise<void>;
  isLoading?: boolean;
  error?: string | null;
  mode: 'create' | 'edit';
}

export const EnvironmentForm: React.FC<EnvironmentFormProps> = ({
  initialData,
  onSubmit,
  isLoading = false,
  error,
  mode,
}) => {
  const navigate = useNavigate();
  const [sshControlEnabled, setSshControlEnabled] = useState(false);
  const [formData, setFormData] = useState<CreateEnvironmentRequest>({
    name: '',
    description: '',
    environmentURL: '',
    target: {
      host: '',
      port: 22,
    },
    credentials: {
      type: 'password',
      username: '',
    },
    healthCheck: {
      enabled: false,
      endpoint: '/health',
      method: 'GET',
      interval: 300,
      timeout: 30,
      validation: {
        type: 'statusCode',
        value: 200,
      },
    },
    commands: {
      type: 'ssh',
      restart: {
        enabled: true,
        command: '',
      },
    },
    upgradeConfig: {
      enabled: false,
      type: 'ssh',
      versionListURL: '',
      jsonPathResponse: '',
      upgradeCommand: {
        command: '',
      },
    },
    metadata: {},
  });

  const [password, setPassword] = useState('');
  const [privateKey, setPrivateKey] = useState('');

  // Initialize form data from initialData if in edit mode
  useEffect(() => {
    if (mode === 'edit' && initialData) {
      // Check if SSH control should be enabled based on credentials or command type
      const hasSshCredentials = !!initialData.credentials.username;
      const hasSshCommands = initialData.commands.type === 'ssh' || initialData.upgradeConfig.type === 'ssh';
      const shouldEnableSsh = hasSshCredentials || hasSshCommands;
      
      setSshControlEnabled(shouldEnableSsh);
      
      setFormData({
        name: initialData.name,
        description: initialData.description,
        environmentURL: initialData.environmentURL || '',
        target: initialData.target,
        credentials: initialData.credentials,
        healthCheck: initialData.healthCheck,
        commands: initialData.commands,
        upgradeConfig: initialData.upgradeConfig,
        metadata: {},
      });
    }
  }, [mode, initialData]);

  // Update command type when SSH control is toggled (only in create mode or when explicitly toggled by user)
  useEffect(() => {
    // Skip this effect during initial load in edit mode
    if (mode === 'edit' && initialData && !formData.name) {
      return;
    }
    
    if (!sshControlEnabled && formData.commands.type === 'ssh') {
      setFormData(prev => ({
        ...prev,
        commands: {
          ...prev.commands,
          type: 'http',
          restart: {
            enabled: prev.commands.restart.enabled,
            url: prev.commands.restart.url || '',
            method: prev.commands.restart.method || 'POST',
            headers: prev.commands.restart.headers,
            body: prev.commands.restart.body,
          }
        },
        upgradeConfig: {
          ...prev.upgradeConfig,
          type: 'http',
        }
      }));
    }
  }, [sshControlEnabled]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Only include SSH fields if SSH control is enabled
    const submitData: CreateEnvironmentRequest = {
      ...formData,
      target: sshControlEnabled ? formData.target : { host: '', port: 22 },
      credentials: sshControlEnabled ? formData.credentials : { type: 'password', username: '' },
      commands: sshControlEnabled ? formData.commands : { 
        ...formData.commands,
        type: 'http', // Force HTTP type when SSH is disabled
        restart: {
          ...formData.commands.restart,
          command: undefined, // Clear SSH command
        }
      },
    };

    await onSubmit(submitData, password, privateKey);
  };

  const handleCancel = () => {
    navigate('/dashboard');
  };

  const handleHttpRestartChange = (data: HttpRequestData) => {
    setFormData({
      ...formData,
      commands: {
        ...formData.commands,
        restart: {
          ...formData.commands.restart,
          url: data.url,
          method: data.method,
          headers: data.headers,
          body: data.body ? (() => {
            try {
              return JSON.parse(data.body);
            } catch {
              return data.body;
            }
          })() : undefined,
        }
      }
    });
  };

  const handleHttpUpgradeChange = (data: HttpRequestData) => {
    setFormData({
      ...formData,
      upgradeConfig: {
        ...formData.upgradeConfig,
        upgradeCommand: {
          ...formData.upgradeConfig.upgradeCommand,
          url: data.url,
          method: data.method,
          headers: data.headers,
          body: data.body ? (() => {
            try {
              return JSON.parse(data.body);
            } catch {
              return data.body;
            }
          })() : undefined,
        }
      }
    });
  };

  const getCommandType = () => {
    if (!sshControlEnabled) {
      return 'http';
    }
    return formData.commands.type;
  };

  return (
    <Box component="form" onSubmit={handleSubmit}>
      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => {}}>
          {error}
        </Alert>
      )}

      {/* Basic Information */}
      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          Basic Information
        </Typography>
        <Divider sx={{ mb: 2 }} />
        
        <Grid container spacing={2}>
          <Grid item xs={12} md={6}>
            <TextField
              label="Environment Name"
              fullWidth
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Production, Staging, Development..."
            />
          </Grid>
          
          <Grid item xs={12} md={6}>
            <TextField
              label="Environment URL"
              fullWidth
              value={formData.environmentURL}
              onChange={(e) => setFormData({ ...formData, environmentURL: e.target.value })}
              placeholder="https://app.example.com"
              helperText="URL to access this environment"
            />
          </Grid>
          
          <Grid item xs={12}>
            <TextField
              label="Description"
              fullWidth
              multiline
              rows={2}
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Brief description of this environment..."
            />
          </Grid>
        </Grid>
      </Paper>

      {/* SSH Control */}
      <Paper sx={{ p: 3, mb: 3 }}>
        <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
          <Typography variant="h6">
            SSH Control
          </Typography>
          <FormControlLabel
            control={
              <Checkbox
                checked={sshControlEnabled}
                onChange={(e) => {
                  const newSshEnabled = e.target.checked;
                  setSshControlEnabled(newSshEnabled);
                  
                  // If disabling SSH, update command types to HTTP
                  if (!newSshEnabled) {
                    setFormData(prev => ({
                      ...prev,
                      commands: {
                        ...prev.commands,
                        type: 'http',
                        restart: {
                          ...prev.commands.restart,
                          url: prev.commands.restart.url || '',
                          method: prev.commands.restart.method || 'POST',
                        }
                      },
                      upgradeConfig: {
                        ...prev.upgradeConfig,
                        type: 'http',
                      }
                    }));
                  }
                }}
              />
            }
            label="Enable SSH Control"
          />
        </Box>
        <Divider sx={{ mb: 2 }} />
        
        {sshControlEnabled ? (
          <Grid container spacing={2}>
            <Grid item xs={12} md={8}>
              <TextField
                label="SSH Host"
                fullWidth
                required
                value={formData.target.host}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  target: { ...formData.target, host: e.target.value }
                })}
                placeholder="192.168.1.100 or example.com"
              />
            </Grid>
            
            <Grid item xs={12} md={4}>
              <TextField
                label="SSH Port"
                type="number"
                fullWidth
                required
                value={formData.target.port}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  target: { ...formData.target, port: parseInt(e.target.value) || 22 }
                })}
              />
            </Grid>
            
            <Grid item xs={12} md={6}>
              <TextField
                label="SSH Username"
                fullWidth
                required
                value={formData.credentials.username}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  credentials: { ...formData.credentials, username: e.target.value }
                })}
              />
            </Grid>
            
            <Grid item xs={12} md={6}>
              <TextField
                select
                label="Authentication Method"
                fullWidth
                required
                value={formData.credentials.type}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  credentials: { ...formData.credentials, type: e.target.value as 'password' | 'key' }
                })}
              >
                <MenuItem value="password">Password</MenuItem>
                <MenuItem value="key">Private Key</MenuItem>
              </TextField>
            </Grid>
            
            {formData.credentials.type === 'password' ? (
              <Grid item xs={12}>
                <TextField
                  label="Password"
                  type="password"
                  fullWidth
                  required={mode === 'create'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  helperText={mode === 'edit' ? "Leave empty to keep current password" : ""}
                />
              </Grid>
            ) : (
              <Grid item xs={12}>
                <TextField
                  label="Private Key"
                  multiline
                  rows={4}
                  fullWidth
                  required={mode === 'create'}
                  value={privateKey}
                  onChange={(e) => setPrivateKey(e.target.value)}
                  placeholder="Paste your private key here..."
                  helperText={mode === 'edit' ? "Leave empty to keep current key" : ""}
                />
              </Grid>
            )}
          </Grid>
        ) : (
          <Typography variant="body2" color="text.secondary">
            SSH control is disabled. Enable it to manage this environment via SSH.
          </Typography>
        )}
      </Paper>

      {/* Health Check Configuration */}
      <Accordion sx={{ mb: 2 }} expanded={formData.healthCheck.enabled}>
        <AccordionSummary 
          expandIcon={<ExpandMore />}
          onClick={(e) => {
            if (e.target === e.currentTarget || (e.target as HTMLElement).closest('.MuiAccordionSummary-expandIconWrapper')) {
              e.preventDefault();
            }
          }}
        >
          <Box display="flex" alignItems="center" gap={2}>
            <Typography variant="h6">Health Check Configuration</Typography>
            <Switch
              checked={formData.healthCheck.enabled}
              onChange={(e) => {
                e.stopPropagation();
                setFormData({ 
                  ...formData, 
                  healthCheck: { ...formData.healthCheck, enabled: e.target.checked }
                });
              }}
              onClick={(e) => e.stopPropagation()}
            />
          </Box>
        </AccordionSummary>
        <AccordionDetails>
          {formData.healthCheck.enabled ? (
            <Grid container spacing={2}>
            <Grid item xs={12} md={8}>
              <TextField
                label="Health Check Endpoint"
                fullWidth
                value={formData.healthCheck.endpoint}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  healthCheck: { ...formData.healthCheck, endpoint: e.target.value }
                })}
                placeholder="/health or https://api.example.com/health"
              />
            </Grid>
            
            <Grid item xs={12} md={4}>
              <TextField
                select
                label="HTTP Method"
                fullWidth
                value={formData.healthCheck.method}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  healthCheck: { ...formData.healthCheck, method: e.target.value }
                })}
              >
                <MenuItem value="GET">GET</MenuItem>
                <MenuItem value="POST">POST</MenuItem>
                <MenuItem value="HEAD">HEAD</MenuItem>
              </TextField>
            </Grid>
            
            <Grid item xs={12} md={4}>
              <TextField
                label="Check Interval (seconds)"
                type="number"
                fullWidth
                value={formData.healthCheck.interval}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  healthCheck: { ...formData.healthCheck, interval: parseInt(e.target.value) || 300 }
                })}
              />
            </Grid>
            
            <Grid item xs={12} md={4}>
              <TextField
                label="Timeout (seconds)"
                type="number"
                fullWidth
                value={formData.healthCheck.timeout}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  healthCheck: { ...formData.healthCheck, timeout: parseInt(e.target.value) || 30 }
                })}
              />
            </Grid>
            
            <Grid item xs={12} md={4}>
              <TextField
                select
                label="Validation Type"
                fullWidth
                value={formData.healthCheck.validation.type}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  healthCheck: { 
                    ...formData.healthCheck, 
                    validation: { 
                      ...formData.healthCheck.validation, 
                      type: e.target.value as 'statusCode' | 'jsonRegex',
                      value: e.target.value === 'statusCode' ? 200 : ''
                    }
                  }
                })}
              >
                <MenuItem value="statusCode">Status Code</MenuItem>
                <MenuItem value="jsonRegex">JSON Regex</MenuItem>
              </TextField>
            </Grid>
            
            <Grid item xs={12}>
              <TextField
                label="Expected Value"
                fullWidth
                value={formData.healthCheck.validation.value}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  healthCheck: { 
                    ...formData.healthCheck, 
                    validation: { 
                      ...formData.healthCheck.validation, 
                      value: formData.healthCheck.validation.type === 'statusCode' 
                        ? parseInt(e.target.value) || 200 
                        : e.target.value
                    }
                  }
                })}
                helperText={
                  formData.healthCheck.validation.type === 'statusCode' 
                    ? "Expected HTTP status code" 
                    : "Regular expression to match in response"
                }
              />
            </Grid>
          </Grid>
          ) : (
            <Typography variant="body2" color="text.secondary">
              Health checks are disabled. Enable to monitor environment health status.
            </Typography>
          )}
        </AccordionDetails>
      </Accordion>

      {/* Commands Configuration */}
      <Accordion sx={{ mb: 2 }} expanded={formData.commands.restart.enabled}>
        <AccordionSummary 
          expandIcon={<ExpandMore />}
          onClick={(e) => {
            if (e.target === e.currentTarget || (e.target as HTMLElement).closest('.MuiAccordionSummary-expandIconWrapper')) {
              e.preventDefault();
            }
          }}
        >
          <Box display="flex" alignItems="center" gap={2}>
            <Typography variant="h6">Restart Configuration</Typography>
            <Switch
              checked={formData.commands.restart.enabled}
              onChange={(e) => {
                e.stopPropagation();
                setFormData({ 
                  ...formData, 
                  commands: { 
                    ...formData.commands, 
                    restart: { ...formData.commands.restart, enabled: e.target.checked }
                  }
                });
              }}
              onClick={(e) => e.stopPropagation()}
            />
          </Box>
        </AccordionSummary>
        <AccordionDetails>
          {formData.commands.restart.enabled ? (
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  select
                  label="Command Type"
                  fullWidth
                  value={getCommandType()}
                  onChange={(e) => setFormData({ 
                    ...formData, 
                    commands: { ...formData.commands, type: e.target.value as 'ssh' | 'http' }
                  })}
                  disabled={!sshControlEnabled}
                >
                  <MenuItem value="ssh" disabled={!sshControlEnabled}>SSH</MenuItem>
                  <MenuItem value="http">HTTP</MenuItem>
                </TextField>
              </Grid>
              
              {getCommandType() === 'http' ? (
                <Grid item xs={12}>
                  <HttpRequestConfig
                    data={{
                      url: formData.commands.restart.url || '',
                      method: formData.commands.restart.method || 'POST',
                      headers: formData.commands.restart.headers,
                      body: typeof formData.commands.restart.body === 'string' 
                        ? formData.commands.restart.body 
                        : JSON.stringify(formData.commands.restart.body || {}),
                    }}
                    onChange={handleHttpRestartChange}
                    urlLabel="Restart Endpoint"
                    urlPlaceholder="http://localhost:8080/restart"
                    urlHelperText="HTTP endpoint to trigger restart"
                  />
                </Grid>
              ) : (
                <Grid item xs={12}>
                  <TextField
                    label="SSH Restart Command"
                    fullWidth
                    value={formData.commands.restart.command || ''}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      commands: { 
                        ...formData.commands, 
                        restart: { ...formData.commands.restart, command: e.target.value }
                      }
                    })}
                    placeholder="sudo systemctl restart app"
                  />
                </Grid>
              )}
            </Grid>
          ) : (
            <Typography variant="body2" color="text.secondary">
              Restart functionality is disabled. Enable it to configure restart commands for this environment.
            </Typography>
          )}
        </AccordionDetails>
      </Accordion>

      {/* Upgrade Configuration */}
      <Accordion sx={{ mb: 3 }} expanded={formData.upgradeConfig.enabled}>
        <AccordionSummary 
          expandIcon={<ExpandMore />}
          onClick={(e) => {
            if (e.target === e.currentTarget || (e.target as HTMLElement).closest('.MuiAccordionSummary-expandIconWrapper')) {
              e.preventDefault();
            }
          }}
        >
          <Box display="flex" alignItems="center" gap={2}>
            <Typography variant="h6">Upgrade Configuration</Typography>
            <Switch
              checked={formData.upgradeConfig.enabled}
              onChange={(e) => {
                e.stopPropagation();
                setFormData({ 
                  ...formData, 
                  upgradeConfig: { ...formData.upgradeConfig, enabled: e.target.checked }
                });
              }}
              onClick={(e) => e.stopPropagation()}
            />
          </Box>
        </AccordionSummary>
        <AccordionDetails>
          {formData.upgradeConfig.enabled && (
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <Typography variant="subtitle1" gutterBottom>
                  Version List Endpoint Configuration
                </Typography>
                <HttpRequestConfig
                  data={{
                    url: formData.upgradeConfig.versionListURL,
                    method: formData.upgradeConfig.versionListMethod || 'GET',
                    headers: formData.upgradeConfig.versionListHeaders,
                    body: formData.upgradeConfig.versionListBody,
                  }}
                  onChange={(data: HttpRequestData) => setFormData({
                    ...formData,
                    upgradeConfig: {
                      ...formData.upgradeConfig,
                      versionListURL: data.url || '',
                      versionListMethod: data.method,
                      versionListHeaders: data.headers,
                      versionListBody: data.body,
                    }
                  })}
                  urlLabel="Version List URL"
                  urlPlaceholder="http://api.example.com/versions"
                  urlHelperText="URL endpoint to fetch available versions"
                />
              </Grid>
              
              <Grid item xs={12}>
                <TextField
                  label="JSONPath Response"
                  fullWidth
                  value={formData.upgradeConfig.jsonPathResponse}
                  onChange={(e) => setFormData({ 
                    ...formData, 
                    upgradeConfig: { ...formData.upgradeConfig, jsonPathResponse: e.target.value }
                  })}
                  placeholder="$.versions[*] or $.data.releases[*]"
                  helperText="JSONPath to extract version list from response (optional)"
                />
              </Grid>
              
              <Grid item xs={12}>
                <Divider sx={{ my: 2 }} />
                <Typography variant="subtitle1" gutterBottom>
                  Upgrade Command Configuration
                </Typography>
              </Grid>
              
              <Grid item xs={12}>
                <TextField
                  select
                  label="Upgrade Command Type"
                  fullWidth
                  value={!sshControlEnabled ? 'http' : formData.upgradeConfig.type}
                  onChange={(e) => setFormData({ 
                    ...formData, 
                    upgradeConfig: { ...formData.upgradeConfig, type: e.target.value as 'ssh' | 'http' }
                  })}
                  disabled={!sshControlEnabled && formData.upgradeConfig.type === 'ssh'}
                >
                  <MenuItem value="ssh" disabled={!sshControlEnabled}>SSH</MenuItem>
                  <MenuItem value="http">HTTP</MenuItem>
                </TextField>
              </Grid>
              
              {(!sshControlEnabled || formData.upgradeConfig.type === 'http') ? (
                <Grid item xs={12}>
                  <HttpRequestConfig
                    data={{
                      url: formData.upgradeConfig.upgradeCommand.url || '',
                      method: formData.upgradeConfig.upgradeCommand.method || 'POST',
                      headers: formData.upgradeConfig.upgradeCommand.headers,
                      body: typeof formData.upgradeConfig.upgradeCommand.body === 'string' 
                        ? formData.upgradeConfig.upgradeCommand.body 
                        : JSON.stringify(formData.upgradeConfig.upgradeCommand.body || {}),
                    }}
                    onChange={handleHttpUpgradeChange}
                    urlLabel="Upgrade Endpoint"
                    urlPlaceholder="http://localhost:8080/upgrade/{VERSION}"
                    urlHelperText="Use {VERSION} as placeholder for version"
                  />
                </Grid>
              ) : (
                <Grid item xs={12}>
                  <TextField
                    label="SSH Upgrade Commands"
                    multiline
                    rows={4}
                    fullWidth
                    value={formData.upgradeConfig.upgradeCommand.command || ''}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      upgradeConfig: { 
                        ...formData.upgradeConfig, 
                        upgradeCommand: { ...formData.upgradeConfig.upgradeCommand, command: e.target.value }
                      }
                    })}
                    placeholder="sudo app-upgrade --version={VERSION}&#10;sudo systemctl restart app"
                    helperText="Use {VERSION} as placeholder. Multiple commands on separate lines."
                  />
                </Grid>
              )}
            </Grid>
          )}
        </AccordionDetails>
      </Accordion>

      {/* Action Buttons */}
      <Box display="flex" gap={2} justifyContent="flex-end">
        <Button
          variant="outlined"
          startIcon={<Cancel />}
          onClick={handleCancel}
          disabled={isLoading}
        >
          Cancel
        </Button>
        <Button
          type="submit"
          variant="contained"
          startIcon={<Save />}
          disabled={isLoading}
        >
          {mode === 'create' ? 'Create Environment' : 'Update Environment'}
        </Button>
      </Box>
    </Box>
  );
};
