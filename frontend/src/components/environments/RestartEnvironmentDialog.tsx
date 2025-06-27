import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
  Box,
  Alert,
  IconButton,
  Paper,
  Stack,
  Chip,
  FormControlLabel,
  Checkbox,
} from '@mui/material';
import {
  Close,
  RestartAlt,
  Info,
  Warning,
  CheckCircle,
} from '@mui/icons-material';
import { Environment } from '@/types/environment';

interface RestartEnvironmentDialogProps {
  open: boolean;
  onClose: () => void;
  environment: Environment;
  onRestart: (force: boolean) => void;
}

export const RestartEnvironmentDialog: React.FC<RestartEnvironmentDialogProps> = ({
  open,
  onClose,
  environment,
  onRestart,
}) => {
  const [confirmationStep, setConfirmationStep] = useState(false);
  const [forceRestart, setForceRestart] = useState(false);

  const handleRestartClick = () => {
    setConfirmationStep(true);
  };

  const handleConfirmRestart = () => {
    onRestart(forceRestart);
    handleClose();
  };

  const handleClose = () => {
    setConfirmationStep(false);
    setForceRestart(false);
    onClose();
  };

  const getCommandPreview = () => {
    const config = environment.commands;
    if (!config) return null;

    if (config.type === 'ssh') {
      return config.restart.command || 'No command configured';
    } else if (config.type === 'http') {
      const method = config.restart.method || 'GET';
      const url = config.restart.url || 'No URL configured';
      
      // If there's a body, show it
      if (config.restart.body && Object.keys(config.restart.body).length > 0) {
        return `${method} ${url}\n\nBody:\n${JSON.stringify(config.restart.body, null, 2)}`;
      }
      
      return `${method} ${url}`;
    }
    return null;
  };

  const isRestartEnabled = environment.commands?.restart?.enabled ?? false;

  return (
    <Dialog 
      open={open} 
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          borderRadius: 2,
        }
      }}
    >
      <DialogTitle sx={{ 
        display: 'flex', 
        alignItems: 'center', 
        justifyContent: 'space-between',
        pb: 2,
      }}>
        <Box display="flex" alignItems="center" gap={1}>
          <RestartAlt sx={{ color: 'primary.main' }} />
          <Typography variant="h6" fontWeight={600}>
            Restart Environment
          </Typography>
        </Box>
        <IconButton 
          onClick={handleClose} 
          size="small"
          sx={{ 
            color: 'text.secondary',
            '&:hover': { bgcolor: 'action.hover' }
          }}
        >
          <Close />
        </IconButton>
      </DialogTitle>

      <DialogContent dividers sx={{ p: 3 }}>
        {!isRestartEnabled ? (
          <Alert severity="warning" icon={<Warning fontSize="inherit" />}>
            <Typography variant="subtitle2" fontWeight={600} gutterBottom>
              Restart Disabled
            </Typography>
            <Typography variant="body2">
              Restart functionality is currently disabled for this environment. 
              Please enable it in the environment configuration to use this feature.
            </Typography>
          </Alert>
        ) : !confirmationStep ? (
          <>
            {/* Environment Info */}
            <Paper 
              elevation={0} 
              sx={{ 
                p: 2, 
                mb: 3, 
                bgcolor: 'background.default',
                border: 1,
                borderColor: 'divider',
              }}
            >
              <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                Environment
              </Typography>
              <Typography variant="h6" fontWeight={600}>
                {environment.name}
              </Typography>
              <Box display="flex" alignItems="center" gap={1} mt={1}>
                <Chip 
                  label={`Status: ${environment.status.health}`}
                  size="small"
                  color={
                    environment.status.health === 'healthy' ? 'success' : 
                    environment.status.health === 'unhealthy' ? 'error' : 
                    'warning'
                  }
                  sx={{ fontWeight: 500 }}
                />
                {environment.timestamps.lastRestartAt && (
                  <Chip 
                    label={`Last restart: ${new Date(environment.timestamps.lastRestartAt).toLocaleDateString()}`}
                    size="small"
                    variant="outlined"
                    sx={{ fontWeight: 500 }}
                  />
                )}
              </Box>
            </Paper>

            {/* Restart Options */}
            <Alert severity="info" icon={<Info fontSize="inherit" />} sx={{ mb: 3 }}>
              <Typography variant="body2">
                This will restart the environment by executing the configured {environment.commands.type === 'ssh' ? 'SSH command' : 'HTTP request'}.
                The environment may be temporarily unavailable during the restart process.
              </Typography>
            </Alert>

            <FormControlLabel
              control={
                <Checkbox
                  checked={forceRestart}
                  onChange={(e) => setForceRestart(e.target.checked)}
                />
              }
              label={
                <Box>
                  <Typography variant="body2" fontWeight={500}>
                    Force Restart
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    Forces the restart even if health checks are failing. Use with caution.
                  </Typography>
                </Box>
              }
            />
          </>
        ) : (
          /* Confirmation Step */
          <Stack spacing={3}>
            <Alert 
              severity={forceRestart ? "error" : "warning"} 
              icon={<Warning fontSize="inherit" />}
            >
              <Typography variant="subtitle2" fontWeight={600} gutterBottom>
                Confirm Restart
              </Typography>
              <Typography variant="body2">
                Are you sure you want to restart <strong>{environment.name}</strong>?
                {forceRestart && (
                  <Box component="span" color="error.main">
                    {' '}This is a force restart and may cause data loss if the environment is in an unstable state.
                  </Box>
                )}
              </Typography>
            </Alert>

            <Paper 
              elevation={0} 
              sx={{ 
                p: 2, 
                bgcolor: 'background.default',
                border: 1,
                borderColor: 'divider',
              }}
            >
              <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                Restart Details
              </Typography>
              <Stack spacing={1.5}>
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    Environment:
                  </Typography>
                  <Typography variant="body2" fontWeight={500}>
                    {environment.name}
                  </Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    Current Status:
                  </Typography>
                  <Typography variant="body2" fontWeight={500}>
                    {environment.status.health}
                  </Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    Restart Type:
                  </Typography>
                  <Typography variant="body2" fontWeight={500}>
                    {forceRestart ? 'Force Restart' : 'Normal Restart'}
                  </Typography>
                </Box>
              </Stack>
            </Paper>

            {getCommandPreview() && (
              <Paper 
                elevation={0} 
                sx={{ 
                  p: 2, 
                  bgcolor: 'grey.900',
                  border: 1,
                  borderColor: 'divider',
                }}
              >
                <Typography variant="subtitle2" color="grey.400" gutterBottom>
                  Command Preview
                </Typography>
                <Typography 
                  variant="body2" 
                  sx={{ 
                    fontFamily: 'monospace',
                    color: 'grey.100',
                    wordBreak: 'break-all',
                  }}
                >
                  {getCommandPreview()}
                </Typography>
              </Paper>
            )}

            <Alert severity="info">
              <Typography variant="body2">
                The restart process will execute the configured {environment.commands.type === 'ssh' ? 'SSH command' : 'HTTP request'}.
                Please ensure all critical data is saved before proceeding.
              </Typography>
            </Alert>
          </Stack>
        )}
      </DialogContent>

      <DialogActions sx={{ p: 2.5, gap: 1 }}>
        {!isRestartEnabled ? (
          <Button onClick={handleClose} color="inherit">
            Close
          </Button>
        ) : !confirmationStep ? (
          <>
            <Button onClick={handleClose} color="inherit">
              Cancel
            </Button>
            <Button 
              onClick={handleRestartClick}
              variant="contained"
              startIcon={<RestartAlt />}
            >
              Continue
            </Button>
          </>
        ) : (
          <>
            <Button 
              onClick={() => setConfirmationStep(false)} 
              color="inherit"
            >
              Back
            </Button>
            <Button 
              onClick={handleConfirmRestart}
              variant="contained"
              color={forceRestart ? "error" : "primary"}
              startIcon={<CheckCircle />}
              sx={{
                bgcolor: forceRestart ? 'error.main' : 'primary.main',
                '&:hover': {
                  bgcolor: forceRestart ? 'error.dark' : 'primary.dark',
                }
              }}
            >
              Confirm Restart
            </Button>
          </>
        )}
      </DialogActions>
    </Dialog>
  );
};
