import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Paper,
  Typography,
  Button,
  Grid,
  CircularProgress,
  Alert,
  Tab,
  Tabs,
  IconButton,
  Chip,
} from '@mui/material';
import {
  ArrowBack,
  Edit,
  Delete,
  Refresh,
  CheckCircle,
  Error,
  Help,
} from '@mui/icons-material';
import { useQuery } from '@tanstack/react-query';
import { format } from 'date-fns';

import { environmentApi } from '@/api/environments';
import { useEnvironmentActions } from '@/hooks/useEnvironmentActions';
import { HealthStatus } from '@/types/environment';
import { EnvironmentLogs } from '@/components/environments';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`tabpanel-${index}`}
      aria-labelledby={`tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

export const EnvironmentDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [tabValue, setTabValue] = React.useState(0);
  const { deleteEnvironment } = useEnvironmentActions();

  const { data: environment, isLoading, error, refetch } = useQuery({
    queryKey: ['environments', id],
    queryFn: () => environmentApi.get(id!),
    enabled: !!id,
    refetchInterval: 5000, // Refresh every 5 seconds
  });

  const handleBack = () => {
    navigate('/dashboard');
  };

  const handleDelete = () => {
    if (id) {
      deleteEnvironment(id);
      navigate('/dashboard');
    }
  };

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="50vh">
        <CircularProgress />
      </Box>
    );
  }

  if (error || !environment) {
    return (
      <Box p={3}>
        <Button startIcon={<ArrowBack />} onClick={handleBack} sx={{ mb: 2 }}>
          Back to Dashboard
        </Button>
        <Alert severity="error">
          {error ? `Failed to load environment: ${(error as any).message}` : 'Environment not found'}
        </Alert>
      </Box>
    );
  }

  const getHealthIcon = () => {
    switch (environment.status.health) {
      case HealthStatus.Healthy:
        return <CheckCircle sx={{ color: 'success.main' }} />;
      case HealthStatus.Unhealthy:
        return <Error sx={{ color: 'error.main' }} />;
      default:
        return <Help sx={{ color: 'warning.main' }} />;
    }
  };

  const getHealthChipColor = () => {
    switch (environment.status.health) {
      case HealthStatus.Healthy:
        return 'success';
      case HealthStatus.Unhealthy:
        return 'error';
      default:
        return 'warning';
    }
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Box display="flex" alignItems="center" gap={2}>
          <IconButton onClick={handleBack}>
            <ArrowBack />
          </IconButton>
          <Typography variant="h4" component="h1">
            {environment.name}
          </Typography>
          <Box display="flex" alignItems="center" gap={1}>
            {getHealthIcon()}
            <Chip
              label={environment.status.health}
              size="small"
              color={getHealthChipColor()}
            />
          </Box>
        </Box>
        <Box display="flex" gap={1}>
          <IconButton onClick={() => refetch()}>
            <Refresh />
          </IconButton>
          <IconButton onClick={() => navigate(`/environments/${id}/edit`)}>
            <Edit />
          </IconButton>
          <IconButton onClick={handleDelete} color="error">
            <Delete />
          </IconButton>
        </Box>
      </Box>

      <Grid container spacing={3}>
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Environment Information
            </Typography>
            <Box sx={{ '& > *': { mb: 2 } }}>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Description
                </Typography>
                <Typography>{environment.description || 'No description'}</Typography>
              </Box>
              {environment.environmentURL && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    Environment URL
                  </Typography>
                  <Typography>
                    <a href={environment.environmentURL} target="_blank" rel="noopener noreferrer">
                      {environment.environmentURL}
                    </a>
                  </Typography>
                </Box>
              )}
              <Box>
                <Typography variant="caption" color="text.secondary">
                  SSH Target
                </Typography>
                <Typography>
                  {environment.target.host}:{environment.target.port}
                </Typography>
              </Box>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  SSH User
                </Typography>
                <Typography>{environment.credentials.username}</Typography>
              </Box>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Auth Type
                </Typography>
                <Typography>{environment.credentials.type}</Typography>
              </Box>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Created
                </Typography>
                <Typography>{format(new Date(environment.timestamps.createdAt), 'MMM d, yyyy, h:mm a')}</Typography>
              </Box>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Last Updated
                </Typography>
                <Typography>{format(new Date(environment.timestamps.updatedAt), 'MMM d, yyyy, h:mm a')}</Typography>
              </Box>
              {environment.timestamps.lastRestartAt && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    Last Restart
                  </Typography>
                  <Typography>{format(new Date(environment.timestamps.lastRestartAt), 'MMM d, yyyy, h:mm a')}</Typography>
                </Box>
              )}
              {environment.timestamps.lastUpgradeAt && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    Last Upgrade
                  </Typography>
                  <Typography>{format(new Date(environment.timestamps.lastUpgradeAt), 'MMM d, yyyy, h:mm a')}</Typography>
                </Box>
              )}
            </Box>
          </Paper>
        </Grid>

        <Grid item xs={12} md={8}>
          <Paper>
            <Tabs value={tabValue} onChange={handleTabChange}>
              <Tab label="System Info" />
              <Tab label="Health Check" />
              <Tab label="Commands" />
              <Tab label="Upgrade Config" />
              <Tab label="Logs" />
            </Tabs>

            <TabPanel value={tabValue} index={0}>
              <Grid container spacing={2}>
                <Grid item xs={12} sm={6}>
                  <Typography variant="caption" color="text.secondary">
                    Application Version
                  </Typography>
                  <Typography>{environment.systemInfo.appVersion || 'Unknown'}</Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="caption" color="text.secondary">
                    OS Version
                  </Typography>
                  <Typography>{environment.systemInfo.osVersion || 'Unknown'}</Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="caption" color="text.secondary">
                    Last System Update
                  </Typography>
                  <Typography>
                    {environment.systemInfo.lastUpdated 
                      ? format(new Date(environment.systemInfo.lastUpdated), 'PPp')
                      : 'Unknown'}
                  </Typography>
                </Grid>
              </Grid>
            </TabPanel>

            <TabPanel value={tabValue} index={1}>
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <Typography variant="caption" color="text.secondary">
                    Status
                  </Typography>
                  <Box display="flex" alignItems="center" gap={1}>
                    <Chip
                      label={environment.status.health}
                      color={getHealthChipColor()}
                    />
                    <Typography variant="body2">
                      {environment.status.message}
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="caption" color="text.secondary">
                    Endpoint
                  </Typography>
                  <Typography>{environment.healthCheck.endpoint}</Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="caption" color="text.secondary">
                    Method
                  </Typography>
                  <Typography>{environment.healthCheck.method}</Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="caption" color="text.secondary">
                    Interval
                  </Typography>
                  <Typography>{environment.healthCheck.interval} seconds</Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="caption" color="text.secondary">
                    Response Time
                  </Typography>
                  <Typography>
                    {environment.status.responseTime > 0
                      ? `${environment.status.responseTime}ms`
                      : 'N/A'}
                  </Typography>
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="caption" color="text.secondary">
                    Last Check
                  </Typography>
                  <Typography>
                    {format(new Date(environment.status.lastCheck), 'PPp')}
                  </Typography>
                </Grid>
              </Grid>
            </TabPanel>

            <TabPanel value={tabValue} index={2}>
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <Typography variant="h6" gutterBottom>
                    Restart Command
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    Type: {environment.commands.type}
                  </Typography>
                </Grid>
                {environment.commands.type === 'ssh' ? (
                  <Grid item xs={12}>
                    <Typography variant="caption" color="text.secondary">
                      SSH Command
                    </Typography>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace', bgcolor: 'grey.100', p: 1, borderRadius: 1 }}>
                      {environment.commands.restart.command || 'No command configured'}
                    </Typography>
                  </Grid>
                ) : (
                  <>
                    <Grid item xs={12}>
                      <Typography variant="caption" color="text.secondary">
                        HTTP Endpoint
                      </Typography>
                      <Typography>{environment.commands.restart.url || 'Not configured'}</Typography>
                    </Grid>
                    <Grid item xs={12} sm={6}>
                      <Typography variant="caption" color="text.secondary">
                        Method
                      </Typography>
                      <Typography>{environment.commands.restart.method || 'POST'}</Typography>
                    </Grid>
                  </>
                )}
              </Grid>
            </TabPanel>

            <TabPanel value={tabValue} index={3}>
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <Typography variant="caption" color="text.secondary">
                    Upgrade Enabled
                  </Typography>
                  <Typography>{environment.upgradeConfig.enabled ? 'Yes' : 'No'}</Typography>
                </Grid>
                {environment.upgradeConfig.enabled && (
                  <>
                    <Grid item xs={12}>
                      <Typography variant="caption" color="text.secondary">
                        Version List URL
                      </Typography>
                      <Typography>{environment.upgradeConfig.versionListURL || 'Not configured'}</Typography>
                    </Grid>
                    <Grid item xs={12}>
                      <Typography variant="caption" color="text.secondary">
                        JSONPath Response
                      </Typography>
                      <Typography sx={{ fontFamily: 'monospace' }}>
                        {environment.upgradeConfig.jsonPathResponse || 'Not configured'}
                      </Typography>
                    </Grid>
                    <Grid item xs={12}>
                      <Typography variant="caption" color="text.secondary">
                        Upgrade Command Type
                      </Typography>
                      <Typography>{environment.upgradeConfig.type}</Typography>
                    </Grid>
                    {environment.upgradeConfig.type === 'ssh' ? (
                      <Grid item xs={12}>
                        <Typography variant="caption" color="text.secondary">
                          SSH Commands
                        </Typography>
                        <Typography variant="body2" sx={{ fontFamily: 'monospace', bgcolor: 'grey.100', p: 1, borderRadius: 1, whiteSpace: 'pre-wrap' }}>
                          {environment.upgradeConfig.upgradeCommand.command || 'No command configured'}
                        </Typography>
                      </Grid>
                    ) : (
                      <>
                        <Grid item xs={12}>
                          <Typography variant="caption" color="text.secondary">
                            HTTP Endpoint
                          </Typography>
                          <Typography>{environment.upgradeConfig.upgradeCommand.url || 'Not configured'}</Typography>
                        </Grid>
                        <Grid item xs={12} sm={6}>
                          <Typography variant="caption" color="text.secondary">
                            Method
                          </Typography>
                          <Typography>{environment.upgradeConfig.upgradeCommand.method || 'POST'}</Typography>
                        </Grid>
                      </>
                    )}
                  </>
                )}
              </Grid>
            </TabPanel>

            <TabPanel value={tabValue} index={4}>
              <EnvironmentLogs environmentId={id!} />
            </TabPanel>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};
