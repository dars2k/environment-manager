import React from 'react';
import {
  Box,
  Grid,
  Typography,
  Button,
  Paper,
  Alert,
  Skeleton,
  alpha,
} from '@mui/material';
import {
  Add,
  CloudQueue,
  CheckCircleOutline,
  ErrorOutline,
  HelpOutline,
  LayersOutlined,
} from '@mui/icons-material';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';

import { environmentApi } from '@/api/environments';
import { EnvironmentCard } from '@/components/environments';

interface StatCardProps {
  value: number;
  label: string;
  icon: React.ReactNode;
  color: string;
}

const StatCard: React.FC<StatCardProps> = ({ value, label, icon, color }) => (
  <Paper
    elevation={1}
    sx={{
      p: 2.5,
      display: 'flex',
      alignItems: 'center',
      gap: 2,
      borderLeft: `3px solid ${color}`,
      position: 'relative',
      overflow: 'hidden',
      transition: 'all 0.2s ease',
      '&:hover': {
        boxShadow: `0 4px 24px rgba(0,0,0,0.5), 0 0 0 1px ${alpha(color, 0.2)}`,
        transform: 'translateY(-2px)',
      },
      '&::before': {
        content: '""',
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        background: `radial-gradient(ellipse at 0% 50%, ${alpha(color, 0.07)} 0%, transparent 60%)`,
        pointerEvents: 'none',
      },
    }}
  >
    <Box
      sx={{
        width: 48,
        height: 48,
        borderRadius: '12px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: alpha(color, 0.12),
        color,
        flexShrink: 0,
      }}
    >
      {icon}
    </Box>
    <Box>
      <Typography
        variant="h4"
        sx={{
          fontFamily: '"Oxanium", sans-serif',
          fontWeight: 700,
          lineHeight: 1,
          color: 'text.primary',
        }}
      >
        {value}
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mt: 0.4 }}>
        {label}
      </Typography>
    </Box>
  </Paper>
);

export const Dashboard: React.FC = () => {
  const navigate = useNavigate();

  const { data, isLoading, error } = useQuery({
    queryKey: ['environments'],
    queryFn: () => environmentApi.list(),
    refetchInterval: 30000,
  });

  const handleCreateNew = () => navigate('/environments/create');

  if (isLoading) {
    return (
      <Box>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={4}>
          <Box>
            <Skeleton variant="text" width={200} height={40} />
            <Skeleton variant="text" width={260} height={24} />
          </Box>
        </Box>
        <Grid container spacing={2.5} mb={4}>
          {[1, 2, 3, 4].map((i) => (
            <Grid item xs={12} sm={6} md={3} key={i}>
              <Skeleton variant="rectangular" height={88} sx={{ borderRadius: '12px' }} />
            </Grid>
          ))}
        </Grid>
        <Grid container spacing={3}>
          {[1, 2, 3].map((i) => (
            <Grid item xs={12} sm={6} lg={4} key={i}>
              <Skeleton variant="rectangular" height={200} sx={{ borderRadius: '12px' }} />
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
  const healthyCount   = environments.filter(e => e.status.health === 'healthy').length;
  const unhealthyCount = environments.filter(e => e.status.health === 'unhealthy').length;
  const unknownCount   = environments.filter(e => e.status.health === 'unknown').length;

  return (
    <Box>
      {/* Header */}
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={4}>
        <Box>
          <Typography
            variant="h4"
            component="h1"
            sx={{ fontFamily: '"Oxanium", sans-serif', fontWeight: 700, mb: 0.5 }}
          >
            Environments
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Monitor and manage your application environments
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<Add />}
          onClick={handleCreateNew}
          size="large"
          sx={{ whiteSpace: 'nowrap' }}
        >
          New Environment
        </Button>
      </Box>

      {/* Stats */}
      <Grid container spacing={2.5} mb={4}>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            value={environments.length}
            label="Total Environments"
            icon={<LayersOutlined fontSize="small" />}
            color="#818cf8"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            value={healthyCount}
            label="Healthy"
            icon={<CheckCircleOutline fontSize="small" />}
            color="#34d399"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            value={unhealthyCount}
            label="Unhealthy"
            icon={<ErrorOutline fontSize="small" />}
            color="#f87171"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            value={unknownCount}
            label="Unknown"
            icon={<HelpOutline fontSize="small" />}
            color="#fbbf24"
          />
        </Grid>
      </Grid>

      {environments.length === 0 ? (
        <Paper
          elevation={1}
          sx={{
            p: 8,
            textAlign: 'center',
            background: 'radial-gradient(ellipse at center, rgba(99,102,241,0.04) 0%, transparent 70%)',
          }}
        >
          <Box
            sx={{
              width: 100,
              height: 100,
              borderRadius: '50%',
              bgcolor: 'rgba(129,140,248,0.08)',
              border: '1px solid rgba(129,140,248,0.15)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto',
              mb: 3,
            }}
          >
            <CloudQueue sx={{ fontSize: 48, color: '#818cf8' }} />
          </Box>
          <Typography variant="h5" fontWeight={600} gutterBottom>
            No environments configured
          </Typography>
          <Typography variant="body2" color="text.secondary" mb={4} sx={{ maxWidth: 380, mx: 'auto' }}>
            Get started by creating your first environment to manage your applications
          </Typography>
          <Button variant="contained" size="large" startIcon={<Add />} onClick={handleCreateNew}>
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
