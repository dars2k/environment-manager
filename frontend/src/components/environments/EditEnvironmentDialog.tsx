import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  FormControlLabel,
  Switch,
  Box,
  Alert,
  MenuItem,
  Grid,
  Typography,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import { ExpandMore } from '@mui/icons-material';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useAppSelector, useAppDispatch } from '@/store';
import { setEnvironmentEditDialogOpen } from '@/store/slices/uiSlice';
import { showSuccess, showError } from '@/store/slices/notificationSlice';
import { environmentApi } from '@/api/environments';
import { Environment, CreateEnvironmentRequest } from '@/types/environment';

interface EditEnvironmentDialogProps {
  environment: Environment | null;
}

export const EditEnvironmentDialog: React.FC<EditEnvironmentDialogProps> = ({ environment }) => {
  const dispatch = useAppDispatch();
  const queryClient = useQueryClient();
  const { environmentEditDialogOpen } = useAppSelector(state => state.ui);

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
      enabled: true,
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
        command: 'sudo systemctl restart app',
      },
    },
    upgradeConfig: {
      enabled: false,
      type: 'ssh',
      versionListURL: '',
      jsonPathResponse: '',
      upgradeCommand: {
        command: 'sudo app-upgrade --version={VERSION}',
      },
    },
    metadata: {},
  });

  const [password, setPassword] = useState('');
  const [privateKey, setPrivateKey] = useState('');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (environment && environmentEditDialogOpen) {
      setFormData({
        name: environment.name,
        description: environment.description,
        environmentURL: environment.environmentURL,
        target: environment.target,
        credentials: environment.credentials,
        healthCheck: environment.healthCheck,
        commands: {
          type: environment.commands?.type || 'ssh',
          restart: {
            enabled: environment.commands?.restart?.enabled ?? true,
            command: environment.commands?.restart?.command || '',
            url: environment.commands?.restart?.url || '',
            method: environment.commands?.restart?.method || 'POST',
            headers: environment.commands?.restart?.headers || {},
            body: environment.commands?.restart?.body || {},
          },
        },
        upgradeConfig: {
          enabled: environment.upgradeConfig?.enabled ?? false,
          type: environment.upgradeConfig?.type || 'ssh',
          versionListURL: environment.upgradeConfig?.versionListURL || '',
          jsonPathResponse: environment.upgradeConfig?.jsonPathResponse || '',
          upgradeCommand: {
            command: environment.upgradeConfig?.upgradeCommand?.command || '',
            url: environment.upgradeConfig?.upgradeCommand?.url || '',
            method: environment.upgradeConfig?.upgradeCommand?.method || 'POST',
            headers: environment.upgradeConfig?.upgradeCommand?.headers || {},
            body: environment.upgradeConfig?.upgradeCommand?.body || {},
          },
        },
        metadata: environment.metadata || {},
      });
    }
  }, [environment, environmentEditDialogOpen]);

  const updateMutation = useMutation({
    mutationFn: async (data: CreateEnvironmentRequest) => {
      if (!environment) return;
      // Add password or key to metadata for backend processing
      const requestData = {
        ...data,
        metadata: {
          ...data.metadata,
          ...(data.credentials.type === 'password' && password ? { password } : {}),
          ...(data.credentials.type === 'key' && privateKey ? { privateKey } : {}),
        },
      };
      return environmentApi.update(environment.id, requestData);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['environments'] });
      queryClient.invalidateQueries({ queryKey: ['environments', environment?.id] });
      dispatch(showSuccess('Environment updated successfully'));
      handleClose();
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || 'Failed to update environment';
      dispatch(showError(errorMessage));
      setError(errorMessage);
    },
  });

  const handleClose = () => {
    dispatch(setEnvironmentEditDialogOpen(false));
    setPassword('');
    setPrivateKey('');
    setError(null);
  };

  const handleSubmit = () => {
    if (!formData.name || !formData.target.host || !formData.credentials.username) {
      setError('Please fill in all required fields');
      return;
    }

    // For updates, password/key is only required if changing authentication type
    // or if they're providing a new value
    updateMutation.mutate(formData);
  };

  if (!environment) {
    return null;
  }

  return (
    <Dialog
      open={environmentEditDialogOpen}
      onClose={handleClose}
      maxWidth="md"
      fullWidth
    >
      <DialogTitle>Edit Environment</DialogTitle>
      <DialogContent>
        <Box sx={{ mt: 2 }}>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
              {error}
            </Alert>
          )}

          <Grid container spacing={2}>
            <Grid item xs={12}>
              <TextField
                label="Name"
                fullWidth
                required
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </Grid>

            <Grid item xs={12}>
              <TextField
                label="Environment URL"
                fullWidth
                value={formData.environmentURL}
                onChange={(e) => setFormData({ ...formData, environmentURL: e.target.value })}
                placeholder="https://app.example.com"
                helperText="URL to access this environment"
              />
            </Grid>

            <Grid item xs={12} sm={8}>
              <TextField
                label="Host"
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

            <Grid item xs={12} sm={4}>
              <TextField
                label="Port"
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

            <Grid item xs={12} sm={6}>
              <TextField
                label="Username"
                fullWidth
                required
                value={formData.credentials.username}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  credentials: { ...formData.credentials, username: e.target.value }
                })}
              />
            </Grid>

            <Grid item xs={12} sm={6}>
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
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  helperText="Leave empty to keep current password"
                />
              </Grid>
            ) : (
              <Grid item xs={12}>
                <TextField
                  label="Private Key"
                  multiline
                  rows={4}
                  fullWidth
                  required
                  value={privateKey}
                  onChange={(e) => setPrivateKey(e.target.value)}
                  placeholder="Paste your private key here..."
                  helperText="Leave empty to keep current key"
                />
              </Grid>
            )}

            <Grid item xs={12}>
              <TextField
                label="Description"
                multiline
                rows={2}
                fullWidth
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />
            </Grid>

            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.healthCheck.enabled}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      healthCheck: { ...formData.healthCheck, enabled: e.target.checked }
                    })}
                  />
                }
                label="Enable Health Checks"
              />
            </Grid>

            {formData.healthCheck.enabled && (
              <>
                <Grid item xs={12}>
                  <TextField
                    label="Health Check Endpoint"
                    fullWidth
                    value={formData.healthCheck.endpoint}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      healthCheck: { ...formData.healthCheck, endpoint: e.target.value }
                    })}
                  />
                </Grid>

                <Grid item xs={12} sm={6}>
                  <TextField
                    label="Health Check Interval (seconds)"
                    type="number"
                    fullWidth
                    value={formData.healthCheck.interval}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      healthCheck: { ...formData.healthCheck, interval: parseInt(e.target.value) || 300 }
                    })}
                  />
                </Grid>

                <Grid item xs={12} sm={6}>
                  <TextField
                    label="Health Check Timeout (seconds)"
                    type="number"
                    fullWidth
                    value={formData.healthCheck.timeout}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      healthCheck: { ...formData.healthCheck, timeout: parseInt(e.target.value) || 30 }
                    })}
                  />
                </Grid>

                <Grid item xs={12} sm={6}>
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

                <Grid item xs={12} sm={6}>
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
                  />
                </Grid>
              </>
            )}
          </Grid>

          {/* Custom Commands Section */}
          <Grid item xs={12}>
            <Accordion defaultExpanded={false}>
              <AccordionSummary expandIcon={<ExpandMore />}>
                <Typography>Custom Commands (Optional)</Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Grid container spacing={2}>
                  <Grid item xs={12}>
                    <TextField
                      select
                      label="Command Type"
                      fullWidth
                      value={formData.commands.type}
                      onChange={(e) => setFormData({ 
                        ...formData, 
                        commands: { ...formData.commands, type: e.target.value as 'ssh' | 'http' }
                      })}
                    >
                      <MenuItem value="ssh">SSH</MenuItem>
                      <MenuItem value="http">HTTP</MenuItem>
                    </TextField>
                  </Grid>

                  <Grid item xs={12}>
                    <Typography variant="subtitle2" gutterBottom>
                      Restart Command
                    </Typography>
                    {formData.commands.type === 'ssh' ? (
                      <TextField
                        label="SSH Command"
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
                    ) : (
                      <>
                        <TextField
                          label="URL"
                          fullWidth
                          value={formData.commands.restart.url || ''}
                          onChange={(e) => setFormData({ 
                            ...formData, 
                            commands: { 
                              ...formData.commands, 
                              restart: { ...formData.commands.restart, url: e.target.value }
                            }
                          })}
                          placeholder="http://localhost:8080/restart"
                          sx={{ mb: 2 }}
                        />
                        <Grid container spacing={2}>
                          <Grid item xs={12} sm={6}>
                            <TextField
                              select
                              label="Method"
                              fullWidth
                              value={formData.commands.restart.method || 'POST'}
                              onChange={(e) => setFormData({ 
                                ...formData, 
                                commands: { 
                                  ...formData.commands, 
                                  restart: { ...formData.commands.restart, method: e.target.value }
                                }
                              })}
                            >
                              <MenuItem value="GET">GET</MenuItem>
                              <MenuItem value="POST">POST</MenuItem>
                              <MenuItem value="PUT">PUT</MenuItem>
                              <MenuItem value="PATCH">PATCH</MenuItem>
                              <MenuItem value="DELETE">DELETE</MenuItem>
                            </TextField>
                          </Grid>
                        </Grid>
                        <TextField
                          label="Headers (Optional)"
                          multiline
                          rows={2}
                          fullWidth
                          sx={{ mt: 2 }}
                          placeholder='{"Authorization": "Bearer token", "Content-Type": "application/json"}'
                          helperText="JSON format for headers"
                          value={JSON.stringify(formData.commands.restart.headers || {})}
                          onChange={(e) => {
                            try {
                              const headers = JSON.parse(e.target.value || '{}');
                              setFormData({ 
                                ...formData, 
                                commands: { 
                                  ...formData.commands, 
                                  restart: { ...formData.commands.restart, headers }
                                }
                              });
                            } catch (err) {
                              // Invalid JSON, don't update
                            }
                          }}
                        />
                        <TextField
                          label="Body (Optional)"
                          multiline
                          rows={3}
                          fullWidth
                          sx={{ mt: 2 }}
                          placeholder='{"action": "restart"}'
                          helperText="JSON format for request body"
                          value={JSON.stringify(formData.commands.restart.body || {})}
                          onChange={(e) => {
                            try {
                              const body = JSON.parse(e.target.value || '{}');
                              setFormData({ 
                                ...formData, 
                                commands: { 
                                  ...formData.commands, 
                                  restart: { ...formData.commands.restart, body }
                                }
                              });
                            } catch (err) {
                              // Invalid JSON, don't update
                            }
                          }}
                        />
                      </>
                    )}
                  </Grid>

                </Grid>
              </AccordionDetails>
            </Accordion>
          </Grid>

          {/* Upgrade Configuration Section */}
          <Grid item xs={12}>
            <Accordion defaultExpanded={false}>
              <AccordionSummary expandIcon={<ExpandMore />}>
                <Typography>Upgrade Configuration (Optional)</Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Grid container spacing={2}>
                  <Grid item xs={12}>
                    <FormControlLabel
                      control={
                        <Switch
                          checked={formData.upgradeConfig.enabled}
                          onChange={(e) => setFormData({ 
                            ...formData, 
                            upgradeConfig: { ...formData.upgradeConfig, enabled: e.target.checked }
                          })}
                        />
                      }
                      label="Enable Version Upgrades"
                    />
                  </Grid>

                  {formData.upgradeConfig.enabled && (
                    <>
                      <Grid item xs={12}>
                        <TextField
                          label="Version List URL"
                          fullWidth
                          value={formData.upgradeConfig.versionListURL}
                          onChange={(e) => setFormData({ 
                            ...formData, 
                            upgradeConfig: { ...formData.upgradeConfig, versionListURL: e.target.value }
                          })}
                          placeholder="http://api.example.com/versions"
                          helperText="URL endpoint to fetch available versions"
                        />
                      </Grid>

                      <Grid item xs={12}>
                        <TextField
                          label="JSONPath Response (Optional)"
                          fullWidth
                          value={formData.upgradeConfig.jsonPathResponse}
                          onChange={(e) => setFormData({ 
                            ...formData, 
                            upgradeConfig: { ...formData.upgradeConfig, jsonPathResponse: e.target.value }
                          })}
                          placeholder="$.versions[*] or $.data.releases[*]"
                          helperText="JSONPath to extract version list from response"
                        />
                      </Grid>

                      <Grid item xs={12}>
                        <TextField
                          select
                          label="Upgrade Command Type"
                          fullWidth
                          value={formData.upgradeConfig.type}
                          onChange={(e) => setFormData({ 
                            ...formData, 
                            upgradeConfig: { ...formData.upgradeConfig, type: e.target.value as 'ssh' | 'http' }
                          })}
                        >
                          <MenuItem value="ssh">SSH</MenuItem>
                          <MenuItem value="http">HTTP</MenuItem>
                        </TextField>
                      </Grid>

                      {formData.upgradeConfig.type === 'ssh' ? (
                        <Grid item xs={12}>
                          <TextField
                            label="SSH Commands"
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
                      ) : (
                        <>
                          <Grid item xs={12}>
                            <TextField
                              label="URL"
                              fullWidth
                              value={formData.upgradeConfig.upgradeCommand.url || ''}
                              onChange={(e) => setFormData({ 
                                ...formData, 
                                upgradeConfig: { 
                                  ...formData.upgradeConfig, 
                                  upgradeCommand: { ...formData.upgradeConfig.upgradeCommand, url: e.target.value }
                                }
                              })}
                              placeholder="http://localhost:8080/upgrade/{VERSION}"
                              helperText="Use {VERSION} as placeholder"
                            />
                          </Grid>
                          <Grid item xs={12} sm={6}>
                            <TextField
                              select
                              label="Method"
                              fullWidth
                              value={formData.upgradeConfig.upgradeCommand.method || 'POST'}
                              onChange={(e) => setFormData({ 
                                ...formData, 
                                upgradeConfig: { 
                                  ...formData.upgradeConfig, 
                                  upgradeCommand: { ...formData.upgradeConfig.upgradeCommand, method: e.target.value }
                                }
                              })}
                            >
                              <MenuItem value="GET">GET</MenuItem>
                              <MenuItem value="POST">POST</MenuItem>
                              <MenuItem value="PUT">PUT</MenuItem>
                              <MenuItem value="PATCH">PATCH</MenuItem>
                              <MenuItem value="DELETE">DELETE</MenuItem>
                            </TextField>
                          </Grid>
                          <Grid item xs={12}>
                            <TextField
                              label="Headers (Optional)"
                              multiline
                              rows={2}
                              fullWidth
                              placeholder='{"Authorization": "Bearer token", "Content-Type": "application/json"}'
                              helperText="JSON format for headers"
                            />
                          </Grid>
                          <Grid item xs={12}>
                            <TextField
                              label="Body (Optional)"
                              multiline
                              rows={3}
                              fullWidth
                              placeholder='{"version": "{VERSION}"}'
                              helperText="JSON format for request body. Use {VERSION} as placeholder."
                            />
                          </Grid>
                        </>
                      )}
                    </>
                  )}
                </Grid>
              </AccordionDetails>
            </Accordion>
          </Grid>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>Cancel</Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          disabled={updateMutation.isPending}
        >
          Update
        </Button>
      </DialogActions>
    </Dialog>
  );
};
