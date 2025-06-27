import React, { useState } from 'react';
import { Box, Typography, Paper } from '@mui/material';
import { useMutation } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { useSnackbar } from 'notistack';

import { EnvironmentForm } from '@/components/environments/EnvironmentForm';
import { environmentApi } from '@/api/environments';
import { CreateEnvironmentRequest } from '@/types/environment';

export const CreateEnvironment: React.FC = () => {
  const navigate = useNavigate();
  const { enqueueSnackbar } = useSnackbar();
  const [error, setError] = useState<string | null>(null);

  const createMutation = useMutation({
    mutationFn: async ({ data, password, privateKey }: { 
      data: CreateEnvironmentRequest; 
      password?: string; 
      privateKey?: string 
    }) => {
      // Add password or key to metadata for backend processing
      const requestData = {
        ...data,
        metadata: {
          ...data.metadata,
          ...(data.credentials.type === 'password' && password ? { password } : {}),
          ...(data.credentials.type === 'key' && privateKey ? { privateKey } : {}),
        },
      };
      return environmentApi.create(requestData);
    },
    onSuccess: () => {
      enqueueSnackbar('Environment created successfully', { variant: 'success' });
      navigate('/dashboard');
    },
    onError: (error: any) => {
      setError(error.response?.data?.message || 'Failed to create environment');
    },
  });

  const handleSubmit = async (data: CreateEnvironmentRequest, password?: string, privateKey?: string) => {
    setError(null);
    await createMutation.mutateAsync({ data, password, privateKey });
  };

  return (
    <Box>
      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Create New Environment
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Configure a new environment for monitoring and management
        </Typography>
      </Paper>

      <EnvironmentForm
        mode="create"
        onSubmit={handleSubmit}
        isLoading={createMutation.isPending}
        error={error}
      />
    </Box>
  );
};
