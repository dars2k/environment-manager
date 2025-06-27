import React, { useState } from 'react';
import { Box, Typography, Paper, CircularProgress, Alert } from '@mui/material';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { useSnackbar } from 'notistack';

import { EnvironmentForm } from '@/components/environments/EnvironmentForm';
import { environmentApi } from '@/api/environments';
import { CreateEnvironmentRequest, UpdateEnvironmentRequest } from '@/types/environment';

export const EditEnvironment: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { enqueueSnackbar } = useSnackbar();
  const [error, setError] = useState<string | null>(null);

  const { data: environment, isLoading: loadingEnvironment } = useQuery({
    queryKey: ['environments', id],
    queryFn: () => environmentApi.get(id!),
    enabled: !!id,
  });

  const updateMutation = useMutation({
    mutationFn: async ({ data, password, privateKey }: { 
      data: CreateEnvironmentRequest; 
      password?: string; 
      privateKey?: string;
    }) => {
      // Only send changed fields
      const updateData: UpdateEnvironmentRequest = {
        name: data.name,
        description: data.description,
        environmentURL: data.environmentURL,
        target: data.target,
        credentials: data.credentials,
        healthCheck: data.healthCheck,
        commands: data.commands,
        upgradeConfig: data.upgradeConfig,
      };

      // Add password or key to metadata only if provided (for updates)
      if (password || privateKey) {
        updateData.metadata = {};
        if (data.credentials.type === 'password' && password) {
          updateData.metadata.password = password;
        } else if (data.credentials.type === 'key' && privateKey) {
          updateData.metadata.privateKey = privateKey;
        }
      }

      return environmentApi.update(id!, updateData);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['environments'] });
      enqueueSnackbar('Environment updated successfully', { variant: 'success' });
      navigate('/dashboard');
    },
    onError: (error: any) => {
      setError(error.response?.data?.message || 'Failed to update environment');
    },
  });

  const handleSubmit = async (data: CreateEnvironmentRequest, password?: string, privateKey?: string) => {
    setError(null);
    await updateMutation.mutateAsync({ data, password, privateKey });
  };

  if (loadingEnvironment) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="50vh">
        <CircularProgress />
      </Box>
    );
  }

  if (!environment) {
    return (
      <Box>
        <Alert severity="error">Environment not found</Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Edit Environment
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Update configuration for {environment.name}
        </Typography>
      </Paper>

      <EnvironmentForm
        mode="edit"
        initialData={environment}
        onSubmit={handleSubmit}
        isLoading={updateMutation.isPending}
        error={error}
      />
    </Box>
  );
};
