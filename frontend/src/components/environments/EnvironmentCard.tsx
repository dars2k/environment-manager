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
} from '@mui/icons-material';
import { format } from 'date-fns';
import { useSnackbar } from 'notistack';

import { Environment, HealthStatus } from '@/types/environment';
import { useEnvironmentActions } from '@/hooks/useEnvironmentActions';
import { UpgradeEnvironmentDialog } from './UpgradeEnvironmentDialog';
import { RestartEnvironmentDialog } from './RestartEnvironmentDialog';

interface EnvironmentCardProps {
  environment: Environment;
}

export const EnvironmentCard: React.FC<EnvironmentCardProps> = ({ environment }) => {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [upgradeDialogOpen, setUpgradeDialogOpen] = useState(false);
  const [restartDialogOpen, setRestartDialogOpen] = useState(false);
  const { enqueueSnackbar } = useSnackbar();
  const { restart, deleteEnvironment, upgrade } = useEnvironmentActions();

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleUpgradeClick = () => {
    if (!environment.upgradeConfig?.enabled || !environment.upgradeConfig?.versionListURL) {
      enqueueSnackbar('Upgrade is not configured for this environment', { variant: 'warning' });
      return;
    }
    handleMenuClose();
    setUpgradeDialogOpen(true);
  };

  const handleUpgradeVersion = (version: string) => {
    upgrade(environment.id, version);
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

  const handleRestart = (force: boolean) => {
    restart(environment.id, force);
  };

  const getHealthIcon = () => {
    switch (environment.status.health) {
      case HealthStatus.Healthy:
        return <CheckCircle sx={{ color: 'success.main', fontSize: 28 }} />;
      case HealthStatus.Unhealthy:
        return <Error sx={{ color: 'error.main', fontSize: 28 }} />;
      default:
        return <Help sx={{ color: 'warning.main', fontSize: 28 }} />;
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
    <Card
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        position: 'relative',
        overflow: 'visible',
        '&:hover .action-button': {
          opacity: 1,
        },
      }}
    >
      {/* Loading indicator */}
      {environment.status.health === HealthStatus.Unknown && environment.healthCheck?.enabled === true && (
        <LinearProgress
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            borderRadius: '16px 16px 0 0',
            height: 3,
          }}
        />
      )}

      <CardContent sx={{ flexGrow: 1, p: 3 }}>
        {/* Header */}
        <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={3}>
          <Box flex={1}>
            <Box display="flex" alignItems="center" gap={1.5} mb={1}>
              {getHealthIcon()}
              <Typography 
                variant="h6" 
                component="h2" 
                fontWeight={600}
                sx={{
                  cursor: 'pointer',
                  '&:hover': {
                    color: 'primary.main',
                    textDecoration: 'underline',
                  },
                }}
                onClick={() => window.location.href = `/environments/${environment.id}`}
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
                  WebkitLineClamp: 2,
                  WebkitBoxOrient: 'vertical',
                }}
              >
                {environment.description}
              </Typography>
            )}
          </Box>
          <IconButton 
            size="small" 
            onClick={(e) => {
              e.stopPropagation();
              handleMenuOpen(e);
            }}
            className="action-button"
            sx={{ 
              opacity: 0,
              transition: 'opacity 0.2s',
              '&:hover': {
                bgcolor: 'rgba(255, 255, 255, 0.08)',
              }
            }}
          >
            <MoreVert />
          </IconButton>
        </Box>

        {/* Status Section */}
        <Stack spacing={2}>
          <Box display="flex" alignItems="center" gap={1}>
            <Chip
              label={environment.status.health}
              size="small"
              color={getHealthChipColor()}
              sx={{ 
                fontWeight: 600,
                color: 'white',
                '& .MuiChip-label': {
                  color: 'white',
                }
              }}
            />
            {environment.status.responseTime > 0 && (
              <Chip
                icon={<AccessTime sx={{ fontSize: 16 }} />}
                label={`${environment.status.responseTime}ms`}
                size="small"
                variant="outlined"
                sx={{ borderColor: 'divider' }}
              />
            )}
          </Box>

          {/* Info Grid */}
          <Box sx={{ display: 'grid', gap: 1.5 }}>
            {environment.environmentURL && (
              <Box display="flex" alignItems="center" gap={1}>
                <LinkIcon sx={{ fontSize: 18, color: 'text.secondary' }} />
                <Typography 
                  variant="body2" 
                  component="a"
                  href={environment.environmentURL}
                  target="_blank"
                  rel="noopener noreferrer"
                  onClick={(e) => e.stopPropagation()}
                  sx={{ 
                    color: 'primary.main',
                    textDecoration: 'none',
                    '&:hover': {
                      textDecoration: 'underline',
                    },
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {environment.environmentURL}
                </Typography>
              </Box>
            )}
            
            <Box display="flex" alignItems="center" gap={1}>
              <Chip
                label={`${environment.systemInfo.appVersion || 'Unknown'}`}
                size="small"
                sx={{ 
                  bgcolor: 'rgba(255, 255, 255, 0.05)',
                  fontWeight: 500,
                  fontSize: '0.75rem',
                }}
              />
            </Box>
          </Box>
        </Stack>
      </CardContent>

      <CardActions 
        sx={{ 
          px: 3, 
          pb: 3, 
          pt: 0,
        }}
      >
        <Typography variant="caption" color="text.secondary">
          Last check: {format(new Date(environment.status.lastCheck), 'MMM d, h:mm a')}
        </Typography>
      </CardActions>

      {/* Context Menu */}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
        PaperProps={{
          sx: {
            minWidth: 180,
          }
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <MenuItem onClick={() => {
          handleMenuClose();
          window.location.href = `/environments/${environment.id}`;
        }}>
          View Details
        </MenuItem>
        <MenuItem onClick={() => {
          handleMenuClose();
          window.location.href = `/environments/${environment.id}/edit`;
        }}>
          <ListItemIcon>
            <Edit fontSize="small" />
          </ListItemIcon>
          <ListItemText>Edit</ListItemText>
        </MenuItem>
        <Divider />
        <MenuItem
          onClick={handleRestartClick}
          disabled={environment.status.health === HealthStatus.Unknown || !environment.commands?.restart?.enabled}
        >
          <ListItemIcon>
            <RestartAlt fontSize="small" sx={{ opacity: environment.commands?.restart?.enabled ? 1 : 0.5 }} />
          </ListItemIcon>
          <ListItemText>
            <Typography variant="inherit" sx={{ opacity: environment.commands?.restart?.enabled ? 1 : 0.5 }}>
              Restart
            </Typography>
          </ListItemText>
        </MenuItem>
        <MenuItem
          onClick={handleUpgradeClick}
          disabled={!environment.upgradeConfig?.enabled || !environment.upgradeConfig?.versionListURL}
        >
          <ListItemIcon>
            <Upgrade fontSize="small" />
          </ListItemIcon>
          <ListItemText>Upgrade</ListItemText>
          {environment.upgradeConfig?.enabled && <ArrowForwardIos fontSize="small" sx={{ ml: 'auto' }} />}
        </MenuItem>
        <Divider />
        <MenuItem
          onClick={() => {
            handleMenuClose();
            deleteEnvironment(environment.id);
          }}
          sx={{ color: 'error.main' }}
        >
          <ListItemIcon>
            <Delete fontSize="small" color="error" />
          </ListItemIcon>
          <ListItemText>Delete</ListItemText>
        </MenuItem>
      </Menu>

      {/* Upgrade Environment Dialog */}
      <UpgradeEnvironmentDialog
        open={upgradeDialogOpen}
        onClose={() => setUpgradeDialogOpen(false)}
        environment={environment}
        onUpgrade={handleUpgradeVersion}
      />

      {/* Restart Environment Dialog */}
      <RestartEnvironmentDialog
        open={restartDialogOpen}
        onClose={() => setRestartDialogOpen(false)}
        environment={environment}
        onRestart={handleRestart}
      />
    </Card>
  );
};
