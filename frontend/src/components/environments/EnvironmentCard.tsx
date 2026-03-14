import React, { useState } from 'react';
import {
  Card,
  CardContent,
  CardActions,
  Typography,
  Box,
  Chip,
  IconButton,
  Menu,
  MenuItem,
  LinearProgress,
  Divider,
  ListItemIcon,
  ListItemText,
  Stack,
  alpha,
} from '@mui/material';
import {
  MoreVert,
  RestartAlt,
  Edit,
  Delete,
  CheckCircle,
  Error,
  Help,
  Upgrade,
  AccessTime,
  Link as LinkIcon,
  ArrowForwardIos,
  OpenInNew,
} from '@mui/icons-material';
import { format } from 'date-fns';
import { useSnackbar } from 'notistack';
import { useNavigate } from 'react-router-dom';

import { Environment, HealthStatus } from '@/types/environment';
import { useEnvironmentActions } from '@/hooks/useEnvironmentActions';
import { UpgradeEnvironmentDialog } from './UpgradeEnvironmentDialog';
import { RestartEnvironmentDialog } from './RestartEnvironmentDialog';

interface EnvironmentCardProps {
  environment: Environment;
}

const STATUS_CONFIG = {
  [HealthStatus.Healthy]: {
    color: '#34d399',
    border: '#34d399',
    bg: 'rgba(52,211,153,0.06)',
    icon: <CheckCircle sx={{ color: '#34d399', fontSize: 22 }} />,
    chipColor: 'success' as const,
  },
  [HealthStatus.Unhealthy]: {
    color: '#f87171',
    border: '#f87171',
    bg: 'rgba(248,113,113,0.06)',
    icon: <Error sx={{ color: '#f87171', fontSize: 22 }} />,
    chipColor: 'error' as const,
  },
  [HealthStatus.Unknown]: {
    color: '#fbbf24',
    border: '#fbbf24',
    bg: 'rgba(251,191,36,0.04)',
    icon: <Help sx={{ color: '#fbbf24', fontSize: 22 }} />,
    chipColor: 'warning' as const,
  },
};

export const EnvironmentCard: React.FC<EnvironmentCardProps> = ({ environment }) => {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [upgradeDialogOpen, setUpgradeDialogOpen] = useState(false);
  const [restartDialogOpen, setRestartDialogOpen] = useState(false);
  const { enqueueSnackbar } = useSnackbar();
  const { restart, deleteEnvironment, upgrade } = useEnvironmentActions();
  const navigate = useNavigate();

  const status = STATUS_CONFIG[environment.status.health] ?? STATUS_CONFIG[HealthStatus.Unknown];

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => setAnchorEl(null);

  const handleCardClick = () => navigate(`/environments/${environment.id}`);

  const handleUpgradeClick = () => {
    if (!environment.upgradeConfig?.enabled || !environment.upgradeConfig?.versionListURL) {
      enqueueSnackbar('Upgrade is not configured for this environment', { variant: 'warning' });
      return;
    }
    handleMenuClose();
    setUpgradeDialogOpen(true);
  };

  const handleRestartClick = () => {
    const isRestartEnabled = environment.commands?.restart?.enabled ?? false;
    if (!isRestartEnabled) {
      enqueueSnackbar('Restart is not configured for this environment', { variant: 'warning' });
      return;
    }
    handleMenuClose();
    setRestartDialogOpen(true);
  };

  return (
    <Card
      onClick={handleCardClick}
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        position: 'relative',
        overflow: 'hidden',
        cursor: 'pointer',
        borderLeft: `3px solid ${status.border}`,
        '&:hover': {
          boxShadow: `0 8px 40px rgba(0,0,0,0.55), 0 0 0 1px ${alpha(status.color, 0.18)}`,
          '& .card-glow': { opacity: 1 },
        },
      }}
    >
      {/* Subtle status glow behind content */}
      <Box
        className="card-glow"
        aria-hidden
        sx={{
          position: 'absolute',
          top: 0, left: 0, right: 0, bottom: 0,
          background: `radial-gradient(ellipse at 0% 0%, ${alpha(status.color, 0.06)} 0%, transparent 60%)`,
          opacity: 0,
          transition: 'opacity 0.3s ease',
          pointerEvents: 'none',
        }}
      />

      {/* Health-check loading bar */}
      {environment.status.health === HealthStatus.Unknown && environment.healthCheck?.enabled === true && (
        <LinearProgress
          sx={{
            position: 'absolute',
            top: 0, left: 0, right: 0,
            height: 2,
            borderRadius: 0,
          }}
        />
      )}

      <CardContent sx={{ flexGrow: 1, p: 2.5 }}>
        {/* Header row */}
        <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
          <Box flex={1} minWidth={0}>
            <Box display="flex" alignItems="center" gap={1} mb={0.5}>
              {status.icon}
              <Typography
                variant="h6"
                component="h2"
                fontWeight={700}
                noWrap
                sx={{
                  fontFamily: '"Oxanium", sans-serif',
                  fontSize: '1rem',
                  color: 'text.primary',
                }}
              >
                {environment.name}
              </Typography>
            </Box>
            {environment.description && (
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  display: '-webkit-box',
                  WebkitLineClamp: 1,
                  WebkitBoxOrient: 'vertical',
                  fontSize: '0.8rem',
                  ml: '30px',
                }}
              >
                {environment.description}
              </Typography>
            )}
          </Box>
          <IconButton
            size="small"
            onClick={handleMenuOpen}
            sx={{ flexShrink: 0, ml: 0.5 }}
          >
            <MoreVert fontSize="small" />
          </IconButton>
        </Box>

        {/* Status chips */}
        <Stack direction="row" spacing={1} flexWrap="wrap" mb={2}>
          <Chip
            label={environment.status.health}
            size="small"
            color={status.chipColor}
            sx={{ fontWeight: 700, fontSize: '0.7rem', letterSpacing: '0.04em' }}
          />
          {environment.status.responseTime > 0 && (
            <Chip
              icon={<AccessTime sx={{ fontSize: '13px !important' }} />}
              label={`${environment.status.responseTime}ms`}
              size="small"
              variant="outlined"
              sx={{ borderColor: 'divider', fontSize: '0.7rem', fontFamily: '"JetBrains Mono", monospace' }}
            />
          )}
        </Stack>

        {/* Info rows */}
        <Stack spacing={1}>
          {environment.environmentURL && (
            <Box
              display="flex"
              alignItems="center"
              gap={1}
              onClick={(e) => e.stopPropagation()}
            >
              <LinkIcon sx={{ fontSize: 15, color: 'text.secondary', flexShrink: 0 }} />
              <Typography
                variant="body2"
                component="a"
                href={environment.environmentURL}
                target="_blank"
                rel="noopener noreferrer"
                sx={{
                  color: '#818cf8',
                  textDecoration: 'none',
                  fontSize: '0.78rem',
                  '&:hover': { textDecoration: 'underline' },
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                {environment.environmentURL}
              </Typography>
              <OpenInNew sx={{ fontSize: 12, color: 'text.secondary', flexShrink: 0, ml: 'auto' }} />
            </Box>
          )}

          {environment.systemInfo?.appVersion && environment.systemInfo.appVersion !== 'Unknown' && (
            <Box display="flex" alignItems="center" gap={1}>
              <Typography
                variant="caption"
                sx={{
                  fontFamily: '"JetBrains Mono", monospace',
                  color: 'text.secondary',
                  bgcolor: 'rgba(129,140,248,0.08)',
                  px: 1,
                  py: 0.25,
                  borderRadius: '4px',
                  fontSize: '0.7rem',
                  letterSpacing: '0.02em',
                }}
              >
                v{environment.systemInfo.appVersion}
              </Typography>
            </Box>
          )}
        </Stack>
      </CardContent>

      <CardActions sx={{ px: 2.5, pb: 2, pt: 0 }}>
        <Typography
          variant="caption"
          color="text.secondary"
          sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.68rem' }}
        >
          Last check: {format(new Date(environment.status.lastCheck), 'MMM d, h:mm a')}
        </Typography>
      </CardActions>

      {/* Context menu */}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
        onClick={(e) => e.stopPropagation()}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
        PaperProps={{ sx: { minWidth: 180 } }}
      >
        <MenuItem onClick={() => { handleMenuClose(); navigate(`/environments/${environment.id}`); }}>
          View Details
        </MenuItem>
        <MenuItem onClick={() => { handleMenuClose(); navigate(`/environments/${environment.id}/edit`); }}>
          <ListItemIcon><Edit fontSize="small" /></ListItemIcon>
          <ListItemText>Edit</ListItemText>
        </MenuItem>
        <Divider />
        <MenuItem
          onClick={handleRestartClick}
          disabled={!environment.commands?.restart?.enabled}
        >
          <ListItemIcon>
            <RestartAlt fontSize="small" sx={{ opacity: environment.commands?.restart?.enabled ? 1 : 0.4 }} />
          </ListItemIcon>
          <ListItemText>
            <Typography variant="inherit" sx={{ opacity: environment.commands?.restart?.enabled ? 1 : 0.4 }}>
              Restart
            </Typography>
          </ListItemText>
        </MenuItem>
        <MenuItem
          onClick={handleUpgradeClick}
          disabled={!environment.upgradeConfig?.enabled || !environment.upgradeConfig?.versionListURL}
        >
          <ListItemIcon><Upgrade fontSize="small" /></ListItemIcon>
          <ListItemText>Upgrade</ListItemText>
          {environment.upgradeConfig?.enabled && <ArrowForwardIos fontSize="small" sx={{ ml: 'auto' }} />}
        </MenuItem>
        <Divider />
        <MenuItem
          onClick={() => { handleMenuClose(); deleteEnvironment(environment.id); }}
          sx={{ color: 'error.main' }}
        >
          <ListItemIcon><Delete fontSize="small" color="error" /></ListItemIcon>
          <ListItemText>Delete</ListItemText>
        </MenuItem>
      </Menu>

      <UpgradeEnvironmentDialog
        open={upgradeDialogOpen}
        onClose={() => setUpgradeDialogOpen(false)}
        environment={environment}
        onUpgrade={(version) => upgrade(environment.id, version)}
      />
      <RestartEnvironmentDialog
        open={restartDialogOpen}
        onClose={() => setRestartDialogOpen(false)}
        environment={environment}
        onRestart={(force) => restart(environment.id, force)}
      />
    </Card>
  );
};
