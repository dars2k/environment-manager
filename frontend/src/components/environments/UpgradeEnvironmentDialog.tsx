import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
  Box,
  CircularProgress,
  Alert,
  MenuItem,
  Chip,
  Paper,
  Stack,
  IconButton,
  Autocomplete,
  TextField,
} from '@mui/material';
import {
  Close,
  Upgrade,
  CheckCircle,
  Info,
} from '@mui/icons-material';
import { useQuery } from '@tanstack/react-query';
import { environmentApi } from '@/api/environments';
import { Environment } from '@/types/environment';

interface UpgradeEnvironmentDialogProps {
  open: boolean;
  onClose: () => void;
  environment: Environment;
  onUpgrade: (version: string) => void;
}

export const UpgradeEnvironmentDialog: React.FC<UpgradeEnvironmentDialogProps> = ({
  open,
  onClose,
  environment,
  onUpgrade,
}) => {
  const [selectedVersion, setSelectedVersion] = useState<string>('');
  const [confirmationStep, setConfirmationStep] = useState(false);

  // Query for available versions
  const { data: versionsData, isLoading, error } = useQuery({
    queryKey: ['versions', environment.id],
    queryFn: () => environmentApi.getVersions(environment.id),
    enabled: open && environment.upgradeConfig?.enabled && Boolean(environment.upgradeConfig?.versionListURL),
  });

  const handleVersionSelect = (version: string) => {
    setSelectedVersion(version);
    setConfirmationStep(false);
  };

  const handleUpgradeClick = () => {
    if (selectedVersion) {
      setConfirmationStep(true);
    }
  };

  const handleConfirmUpgrade = () => {
    onUpgrade(selectedVersion);
    handleClose();
  };

  const handleClose = () => {
    setSelectedVersion('');
    setConfirmationStep(false);
    onClose();
  };

  const getCommandPreview = () => {
    const config = environment.upgradeConfig;
    if (!config || !selectedVersion) return null;

    if (config.type === 'ssh') {
      return config.upgradeCommand.command?.replace('{VERSION}', selectedVersion);
    } else if (config.type === 'http') {
      const method = config.upgradeCommand.method || 'GET';
      const url = config.upgradeCommand.url?.replace('{VERSION}', selectedVersion);
      
      // If there's a body, show it with replaced version
      if (config.upgradeCommand.body && Object.keys(config.upgradeCommand.body).length > 0) {
        const replacedBody: Record<string, any> = {};
        Object.entries(config.upgradeCommand.body).forEach(([key, value]) => {
          if (typeof value === 'string') {
            replacedBody[key] = value.replace('{VERSION}', selectedVersion);
          } else {
            replacedBody[key] = value;
          }
        });
        return `${method} ${url}\n\nBody:\n${JSON.stringify(replacedBody, null, 2)}`;
      }
      
      return `${method} ${url}`;
    }
    return null;
  };

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
          <Upgrade sx={{ color: 'primary.main' }} />
          <Typography variant="h6" fontWeight={600}>
            Upgrade Environment
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
        {!confirmationStep ? (
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
                  label={`Current version: ${versionsData?.currentVersion || environment.systemInfo.appVersion || 'Unknown'}`}
                  size="small"
                  sx={{ fontWeight: 500 }}
                />
              </Box>
            </Paper>

            {/* Version Selection */}
            {isLoading ? (
              <Box textAlign="center" py={4}>
                <CircularProgress size={40} />
                <Typography variant="body2" color="text.secondary" mt={2}>
                  Loading available versions...
                </Typography>
              </Box>
            ) : error ? (
              <Alert severity="error" sx={{ mb: 3 }}>
                Failed to load available versions. Please try again.
              </Alert>
            ) : versionsData?.availableVersions && versionsData.availableVersions.length > 0 ? (
              <>
                <Autocomplete
                  fullWidth
                  value={selectedVersion}
                  onChange={(_, newValue) => handleVersionSelect(newValue || '')}
                  options={versionsData.availableVersions}
                  getOptionDisabled={(option) => option === versionsData.currentVersion}
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      label="Select Version"
                      placeholder="Type to search or select a version"
                      sx={{ mb: 3 }}
                    />
                  )}
                  renderOption={(props, option) => (
                    <MenuItem {...props} key={option}>
                      <Box display="flex" alignItems="center" justifyContent="space-between" width="100%">
                        <Typography>{option}</Typography>
                        {option === versionsData.currentVersion && (
                          <Chip label="Current" size="small" color="primary" sx={{ ml: 2 }} />
                        )}
                      </Box>
                    </MenuItem>
                  )}
                  slotProps={{
                    popper: {
                      placement: 'bottom-start',
                      modifiers: [
                        {
                          name: 'flip',
                          enabled: false,
                        },
                      ],
                    },
                  }}
                />

                {selectedVersion && (
                  <Alert severity="info" icon={<Info fontSize="inherit" />}>
                    <Typography variant="body2">
                      You are about to upgrade from version <strong>{versionsData.currentVersion || 'Unknown'}</strong> to version <strong>{selectedVersion}</strong>.
                      This operation will update the application to the selected version.
                    </Typography>
                  </Alert>
                )}
              </>
            ) : (
              <Alert severity="warning">
                No upgrade versions available at this time.
              </Alert>
            )}
          </>
        ) : (
          /* Confirmation Step */
          <Stack spacing={3}>
            <Alert severity="warning" icon={<Info fontSize="inherit" />}>
              <Typography variant="subtitle2" fontWeight={600} gutterBottom>
                Confirm Upgrade
              </Typography>
              <Typography variant="body2">
                Are you sure you want to upgrade <strong>{environment.name}</strong> from version <strong>{versionsData?.currentVersion || 'Unknown'}</strong> to version <strong>{selectedVersion}</strong>?
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
                Upgrade Details
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
                    Current Version:
                  </Typography>
                  <Typography variant="body2" fontWeight={500}>
                    {versionsData?.currentVersion || environment.systemInfo.appVersion || 'Unknown'}
                  </Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    Target Version:
                  </Typography>
                  <Typography variant="body2" fontWeight={500} color="primary.main">
                    {selectedVersion}
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
                The upgrade process will execute the configured {environment.upgradeConfig?.type === 'ssh' ? 'SSH command' : 'HTTP request'} with the selected version.
                This may cause temporary downtime during the upgrade.
              </Typography>
            </Alert>
          </Stack>
        )}
      </DialogContent>

      <DialogActions sx={{ p: 2.5, gap: 1 }}>
        {!confirmationStep ? (
          <>
            <Button onClick={handleClose} color="inherit">
              Cancel
            </Button>
            <Button 
              onClick={handleUpgradeClick}
              variant="contained"
              disabled={!selectedVersion || isLoading}
              startIcon={<Upgrade />}
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
              onClick={handleConfirmUpgrade}
              variant="contained"
              color="primary"
              startIcon={<CheckCircle />}
              sx={{
                bgcolor: 'primary.main',
                '&:hover': {
                  bgcolor: 'primary.dark',
                }
              }}
            >
              Confirm Upgrade
            </Button>
          </>
        )}
      </DialogActions>
    </Dialog>
  );
};
