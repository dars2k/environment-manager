import React from 'react';
import {
  Box,
  Grid,
  Typography,
  Button,
  Paper,
  Alert,
  Skeleton,
  Stack,
  useTheme,
  alpha,
} from '@mui/material';
import { Add, CloudQueue } from '@mui/icons-material';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';

import { environmentApi } from '@/api/environments';
import { EnvironmentCard } from '@/components/environments';

export const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const theme = useTheme();

  const { data, isLoading, error } = useQuery({
    queryKey: ['environments'],
    queryFn: () => environmentApi.list(),
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  const handleCreateNew = () => {
    navigate('/environments/create');
  };

  if (isLoading) {
    return (
      <Box>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={4}>
          <Box>
            <Typography variant="h4" component="h1" fontWeight={600} gutterBottom>
              Environments
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Manage your application environments
            </Typography>
          </Box>
        </Box>
        <Grid container spacing={3}>
          {[1, 2, 3, 4].map((i) => (
            <Grid item xs={12} sm={6} lg={4} key={i}>
              <Paper sx={{ p: 3, height: 300 }}>
                <Skeleton variant="circular" width={28} height={28} sx={{ mb: 2 }} />
                <Skeleton variant="text" width="60%" height={32} sx={{ mb: 1 }} />
                <Skeleton variant="text" width="100%" />
                <Skeleton variant="text" width="80%" sx={{ mb: 3 }} />
                <Stack spacing={1}>
                  <Skeleton variant="rectangular" height={30} />
                  <Skeleton variant="rectangular" height={30} />
                  <Skeleton variant="rectangular" height={30} />
                </Stack>
              </Paper>
            </Grid>
          ))}
        </Grid>
      </Box>
    );
  }

  if (error) {
    return (
      <Box p={3}>
        <Alert severity="error">
          Failed to load environments: {(error as any).message}
        </Alert>
      </Box>
    );
  }

  const environments = data?.environments || [];

  return (
    <Box>
      {/* Header Section */}
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={4}>
        <Box>
          <Typography variant="h4" component="h1" fontWeight={600} gutterBottom>
            Environments
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Manage your application environments
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<Add />}
          onClick={handleCreateNew}
          size="large"
        >
          New Environment
        </Button>
      </Box>

      {/* Stats Bar */}
      <Grid container spacing={3} mb={4}>
        <Grid item xs={12} sm={6} md={3}>
          <Paper 
            sx={{ 
              p: 3,
              background: `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.1)} 0%, ${alpha(theme.palette.primary.dark, 0.05)} 100%)`,
              borderColor: alpha(theme.palette.primary.main, 0.2),
            }}
          >
            <Stack spacing={1}>
              <Typography variant="h3" fontWeight={700}>
                {environments.length}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Total Environments
              </Typography>
            </Stack>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Paper 
            sx={{ 
              p: 3,
              background: `linear-gradient(135deg, ${alpha(theme.palette.success.main, 0.1)} 0%, ${alpha(theme.palette.success.dark, 0.05)} 100%)`,
              borderColor: alpha(theme.palette.success.main, 0.2),
            }}
          >
            <Stack spacing={1}>
              <Typography variant="h3" fontWeight={700}>
                {environments.filter(env => env.status.health === 'healthy').length}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Healthy
              </Typography>
            </Stack>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Paper 
            sx={{ 
              p: 3,
              background: `linear-gradient(135deg, ${alpha(theme.palette.error.main, 0.1)} 0%, ${alpha(theme.palette.error.dark, 0.05)} 100%)`,
              borderColor: alpha(theme.palette.error.main, 0.2),
            }}
          >
            <Stack spacing={1}>
              <Typography variant="h3" fontWeight={700}>
                {environments.filter(env => env.status.health === 'unhealthy').length}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Unhealthy
              </Typography>
            </Stack>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Paper 
            sx={{ 
              p: 3,
              background: `linear-gradient(135deg, ${alpha(theme.palette.warning.main, 0.1)} 0%, ${alpha(theme.palette.warning.dark, 0.05)} 100%)`,
              borderColor: alpha(theme.palette.warning.main, 0.2),
            }}
          >
            <Stack spacing={1}>
              <Typography variant="h3" fontWeight={700}>
                {environments.filter(env => env.status.health === 'unknown').length}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Unknown
              </Typography>
            </Stack>
          </Paper>
        </Grid>
      </Grid>

      {environments.length === 0 ? (
        <Paper 
          sx={{ 
            p: 8, 
            textAlign: 'center',
            background: `radial-gradient(circle at center, ${alpha(theme.palette.primary.main, 0.03)} 0%, transparent 70%)`,
          }}
        >
          <Box
            sx={{
              width: 120,
              height: 120,
              borderRadius: '50%',
              bgcolor: alpha(theme.palette.primary.main, 0.1),
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto',
              mb: 3,
            }}
          >
            <CloudQueue sx={{ fontSize: 60, color: 'primary.main' }} />
          </Box>
          <Typography variant="h5" fontWeight={600} gutterBottom>
            No environments configured
          </Typography>
          <Typography variant="body1" color="text.secondary" mb={4} sx={{ maxWidth: 400, mx: 'auto' }}>
            Get started by creating your first environment to manage your applications
          </Typography>
          <Button
            variant="contained"
            size="large"
            startIcon={<Add />}
            onClick={handleCreateNew}
          >
            Create Your First Environment
          </Button>
        </Paper>
      ) : (
        <Grid container spacing={3}>
          {environments.map((env) => (
            <Grid item xs={12} md={6} lg={4} key={env.id}>
              <EnvironmentCard environment={env} />
            </Grid>
          ))}
        </Grid>
      )}
    </Box>
  );
};
