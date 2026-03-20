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
  Tooltip,
  Divider,
  Link,
} from '@mui/material';
import {
  ArrowBack,
  Edit,
  Delete,
  Refresh,
  CheckCircle,
  Error,
  Help,
  OpenInNew,
  InfoOutlined,
} from '@mui/icons-material';
import { useQuery } from '@tanstack/react-query';
import { format, formatDistanceToNow } from 'date-fns';

import { environmentApi } from '@/api/environments';
import { useEnvironmentActions } from '@/hooks/useEnvironmentActions';
import { HealthStatus } from '@/types/environment';
import { EnvironmentLogs } from '@/components/environments';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel({ children, value, index, ...other }: TabPanelProps) {
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

/** Label + value pair used throughout the info panel and tabs */
function InfoField({
  label,
  children,
  mono = false,
}: {
  label: string;
  children: React.ReactNode;
  mono?: boolean;
}) {
  return (
    <Box>
      <Typography variant="caption" color="text.secondary" display="block" gutterBottom={false}>
        {label}
      </Typography>
      <Typography
        variant="body2"
        component="div"
        sx={mono ? { fontFamily: '"JetBrains Mono", monospace', fontSize: '0.8rem' } : undefined}
      >
        {children}
      </Typography>
    </Box>
  );
}

/** Absolute date with relative time on hover */
function TimestampField({ label, value }: { label: string; value: string | null | undefined }) {
  if (!value) return null;
  const date = new Date(value);
  const abs = format(date, 'MMM d, yyyy, h:mm a');
  const rel = formatDistanceToNow(date, { addSuffix: true });
  return (
    <InfoField label={label}>
      <Tooltip title={rel} placement="right">
        <span style={{ cursor: 'default', borderBottom: '1px dotted' }}>{abs}</span>
      </Tooltip>
    </InfoField>
  );
}

/** Section divider with a label inside the info panel */
function SectionDivider({ label }: { label: string }) {
  return (
    <Box sx={{ pt: 0.5, pb: 0.5 }}>
      <Divider>
        <Typography variant="caption" color="text.disabled" sx={{ px: 1, letterSpacing: '0.08em', textTransform: 'uppercase', fontSize: '0.6rem' }}>
          {label}
        </Typography>
      </Divider>
    </Box>
  );
}

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

export const EnvironmentDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [tabValue, setTabValue] = React.useState(0);
  const { deleteEnvironment } = useEnvironmentActions();

  const { data: environment, isLoading, error, refetch } = useQuery({
    queryKey: ['environments', id],
    queryFn: () => environmentApi.get(id!),
    enabled: !!id,
    refetchInterval: 5000,
  });

  const handleBack = () => navigate('/dashboard');

  const handleDelete = () => {
    if (id) {
      deleteEnvironment(id);
      navigate('/dashboard');
    }
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

  let healthIcon: React.ReactNode = <Help sx={{ color: 'warning.main' }} />;
  if (environment.status.health === HealthStatus.Healthy) {
    healthIcon = <CheckCircle sx={{ color: 'success.main' }} />;
  } else if (environment.status.health === HealthStatus.Unhealthy) {
    healthIcon = <Error sx={{ color: 'error.main' }} />;
  }

  let healthChipColor: 'success' | 'error' | 'warning' = 'warning';
  if (environment.status.health === HealthStatus.Healthy) {
    healthChipColor = 'success';
  } else if (environment.status.health === HealthStatus.Unhealthy) {
    healthChipColor = 'error';
  }

  return (
    <Box>
      {/* ── Page header ─────────────────────────────────────────────────── */}
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Box display="flex" alignItems="center" gap={2}>
          <Tooltip title="Back to Dashboard">
            <IconButton onClick={handleBack}>
              <ArrowBack />
            </IconButton>
          </Tooltip>
          <Typography variant="h4" component="h1">
            {environment.name}
          </Typography>
          <Box display="flex" alignItems="center" gap={1}>
            {healthIcon}
            <Chip label={environment.status.health} size="small" color={healthChipColor} />
          </Box>
        </Box>
        <Box display="flex" gap={1}>
          <Tooltip title="Refresh">
            <IconButton onClick={() => refetch()}>
              <Refresh />
            </IconButton>
          </Tooltip>
          <Tooltip title="Edit environment">
            <IconButton onClick={() => navigate(`/environments/${id}/edit`)}>
              <Edit />
            </IconButton>
          </Tooltip>
          <Tooltip title="Delete environment">
            <IconButton onClick={handleDelete} color="error">
              <Delete />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      <Grid container spacing={3}>
        {/* ── Left info panel ─────────────────────────────────────────── */}
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Environment Information
            </Typography>

            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <InfoField label="Description">
                {environment.description || 'No description'}
              </InfoField>

              <SectionDivider label="Connection" />

              {environment.environmentURL && (
                <InfoField label="Environment URL">
                  <Link
                    href={environment.environmentURL}
                    target="_blank"
                    rel="noopener noreferrer"
                    underline="hover"
                    sx={{ display: 'inline-flex', alignItems: 'center', gap: 0.5 }}
                  >
                    {environment.environmentURL}
                    <OpenInNew sx={{ fontSize: 13 }} />
                  </Link>
                </InfoField>
              )}

              <InfoField label="SSH Target">
                {environment.target.host}:{environment.target.port}
              </InfoField>

              <InfoField label="SSH User">
                {environment.credentials.username}
              </InfoField>

              <InfoField label="Auth Type">
                <Chip
                  label={environment.credentials.type}
                  size="small"
                  variant="outlined"
                  sx={{ textTransform: 'capitalize' }}
                />
              </InfoField>

              <SectionDivider label="Timeline" />

              <TimestampField label="Created"      value={environment.timestamps.createdAt} />
              <TimestampField label="Last Updated" value={environment.timestamps.updatedAt} />
              <TimestampField label="Last Restart" value={environment.timestamps.lastRestartAt} />
              <TimestampField label="Last Upgrade" value={environment.timestamps.lastUpgradeAt} />
            </Box>
          </Paper>
        </Grid>

        {/* ── Tabs panel ───────────────────────────────────────────────── */}
        <Grid item xs={12} md={8}>
          <Paper>
            <Tabs value={tabValue} onChange={(_e, v) => setTabValue(v)}>
              <Tab label="System Info" />
              <Tab label="Health Check" />
              <Tab label="Commands" />
              <Tab label="Upgrade Config" />
              <Tab label="Logs" />
            </Tabs>

            {/* System Info */}
            <TabPanel value={tabValue} index={0}>
              <Grid container spacing={2}>
                <Grid item xs={12} sm={6}>
                  <InfoField label="Application Version">
                    {environment.systemInfo.appVersion
                      ? <Typography variant="body2" sx={{ fontFamily: '"JetBrains Mono", monospace' }}>v{environment.systemInfo.appVersion}</Typography>
                      : <Typography variant="body2" color="text.disabled">Not available</Typography>}
                  </InfoField>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <InfoField label="OS Version">
                    {environment.systemInfo.osVersion || <Typography variant="body2" color="text.disabled">Not available</Typography>}
                  </InfoField>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <InfoField label="Last System Update">
                    {environment.systemInfo.lastUpdated
                      ? format(new Date(environment.systemInfo.lastUpdated), 'PPp')
                      : <Typography variant="body2" color="text.disabled">Not available</Typography>}
                  </InfoField>
                </Grid>
              </Grid>
              <Alert severity="info" icon={<InfoOutlined />} sx={{ mt: 3 }}>
                System info is populated automatically when the environment reports in. Values show <em>Not available</em> until the first successful health check.
              </Alert>
            </TabPanel>

            {/* Health Check */}
            <TabPanel value={tabValue} index={1}>
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <InfoField label="Status">
                    <Box display="flex" alignItems="center" gap={1} mt={0.5}>
                      <Chip label={environment.status.health} color={healthChipColor} size="small" />
                      {environment.status.message && (
                        <Typography variant="body2" color="text.secondary">
                          {environment.status.message}
                        </Typography>
                      )}
                    </Box>
                  </InfoField>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <InfoField label="Endpoint">{environment.healthCheck.endpoint}</InfoField>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <InfoField label="Method">{environment.healthCheck.method}</InfoField>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <InfoField label="Interval">{environment.healthCheck.interval} seconds</InfoField>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <InfoField label="Response Time">
                    {environment.status.responseTime > 0
                      ? `${environment.status.responseTime}ms`
                      : 'N/A'}
                  </InfoField>
                </Grid>
                <Grid item xs={12}>
                  <InfoField label="Last Check">
                    <Tooltip title={formatDistanceToNow(new Date(environment.status.lastCheck), { addSuffix: true })} placement="right">
                      <span style={{ cursor: 'default', borderBottom: '1px dotted' }}>
                        {format(new Date(environment.status.lastCheck), 'PPp')}
                      </span>
                    </Tooltip>
                  </InfoField>
                </Grid>
              </Grid>
            </TabPanel>

            {/* Commands */}
            <TabPanel value={tabValue} index={2}>
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <Box display="flex" alignItems="center" gap={1} mb={1}>
                    <Typography variant="h6">Restart Command</Typography>
                    <Chip label={environment.commands.type} size="small" variant="outlined" sx={{ textTransform: 'uppercase' }} />
                  </Box>
                </Grid>
                {environment.commands.type === 'ssh' ? (
                  <Grid item xs={12}>
                    <InfoField label="SSH Command" mono>
                      <Box
                        component="pre"
                        sx={{
                          fontFamily: '"JetBrains Mono", monospace',
                          fontSize: '0.8rem',
                          bgcolor: 'action.selected',
                          border: '1px solid',
                          borderColor: 'divider',
                          borderRadius: 1,
                          p: 1.5,
                          m: 0,
                          whiteSpace: 'pre-wrap',
                          wordBreak: 'break-all',
                        }}
                      >
                        {environment.commands.restart.command || 'No command configured'}
                      </Box>
                    </InfoField>
                  </Grid>
                ) : (
                  <>
                    <Grid item xs={12}>
                      <InfoField label="HTTP Endpoint">
                        {environment.commands.restart.url || (
                          <Typography variant="body2" color="text.disabled">Not configured</Typography>
                        )}
                      </InfoField>
                    </Grid>
                    <Grid item xs={12} sm={6}>
                      <InfoField label="Method">
                        {environment.commands.restart.method || 'POST'}
                      </InfoField>
                    </Grid>
                  </>
                )}
                {!environment.commands.restart.command && !environment.commands.restart.url && (
                  <Grid item xs={12}>
                    <Alert severity="warning">
                      No restart command configured. Add one in the environment settings.
                    </Alert>
                  </Grid>
                )}
              </Grid>
            </TabPanel>

            {/* Upgrade Config */}
            <TabPanel value={tabValue} index={3}>
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <InfoField label="Upgrade Enabled">
                    <Chip
                      label={environment.upgradeConfig.enabled ? 'Enabled' : 'Disabled'}
                      size="small"
                      color={environment.upgradeConfig.enabled ? 'success' : 'default'}
                      variant={environment.upgradeConfig.enabled ? 'filled' : 'outlined'}
                    />
                  </InfoField>
                </Grid>

                {environment.upgradeConfig.enabled && (
                  <>
                    <Grid item xs={12}>
                      <InfoField label="Version List URL">
                        {environment.upgradeConfig.versionListURL || (
                          <Typography variant="body2" color="text.disabled">Not configured</Typography>
                        )}
                      </InfoField>
                      <Alert severity="info" icon={false} sx={{ mt: 1, py: 0.5 }}>
                        <Typography variant="caption">
                          Use <code style={{ fontFamily: 'monospace' }}>{'{VERSION}'}</code> in your upgrade command to insert the selected version at runtime.
                        </Typography>
                      </Alert>
                    </Grid>

                    <Grid item xs={12}>
                      <InfoField label="JSONPath Response" mono>
                        {environment.upgradeConfig.jsonPathResponse || (
                          <Typography variant="body2" color="text.disabled">Not configured</Typography>
                        )}
                      </InfoField>
                    </Grid>

                    <Grid item xs={12}>
                      <Box display="flex" alignItems="center" gap={1} mb={1} mt={1}>
                        <Typography variant="h6">Upgrade Command</Typography>
                        <Chip label={environment.upgradeConfig.type} size="small" variant="outlined" sx={{ textTransform: 'uppercase' }} />
                      </Box>
                    </Grid>

                    {environment.upgradeConfig.type === 'ssh' ? (
                      <Grid item xs={12}>
                        <InfoField label="SSH Command">
                          <Box
                            component="pre"
                            sx={{
                              fontFamily: '"JetBrains Mono", monospace',
                              fontSize: '0.8rem',
                              bgcolor: 'action.selected',
                              border: '1px solid',
                              borderColor: 'divider',
                              borderRadius: 1,
                              p: 1.5,
                              m: 0,
                              whiteSpace: 'pre-wrap',
                              wordBreak: 'break-all',
                            }}
                          >
                            {environment.upgradeConfig.upgradeCommand.command || 'No command configured'}
                          </Box>
                        </InfoField>
                      </Grid>
                    ) : (
                      <>
                        <Grid item xs={12}>
                          <InfoField label="HTTP Endpoint">
                            {environment.upgradeConfig.upgradeCommand.url || (
                              <Typography variant="body2" color="text.disabled">Not configured</Typography>
                            )}
                          </InfoField>
                        </Grid>
                        <Grid item xs={12} sm={6}>
                          <InfoField label="Method">
                            {environment.upgradeConfig.upgradeCommand.method || 'POST'}
                          </InfoField>
                        </Grid>
                      </>
                    )}
                  </>
                )}
              </Grid>
            </TabPanel>

            {/* Logs */}
            <TabPanel value={tabValue} index={4}>
              <EnvironmentLogs environmentId={id!} />
            </TabPanel>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};
