import React from 'react';
import {
  Box,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  CircularProgress,
  Alert,
  Typography,
} from '@mui/material';
import { CheckCircle, Error, Warning, Info } from '@mui/icons-material';
import { format } from 'date-fns';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

interface LogEntry {
  id: string;
  timestamp: string;
  type: 'health_check' | 'action' | 'system' | 'error' | 'auth';
  level: 'info' | 'warning' | 'error' | 'success';
  action?: string;
  message: string;
  details?: any;
}

interface EnvironmentLogsProps {
  environmentId: string;
}

export const EnvironmentLogs: React.FC<EnvironmentLogsProps> = ({ environmentId }) => {
  const { data: logs, isLoading, error } = useQuery({
    queryKey: ['environment-logs', environmentId],
    queryFn: async () => {
      const response = await axios.get(`/api/v1/environments/${environmentId}/logs?limit=50`);
      return response.data.data.logs as LogEntry[];
    },
    refetchInterval: 10000, // Refresh every 10 seconds
  });

  const getLogIcon = (level: string) => {
    switch (level) {
      case 'success':
        return <CheckCircle sx={{ color: 'success.main', fontSize: 16 }} />;
      case 'error':
        return <Error sx={{ color: 'error.main', fontSize: 16 }} />;
      case 'warning':
        return <Warning sx={{ color: 'warning.main', fontSize: 16 }} />;
      default:
        return <Info sx={{ color: 'info.main', fontSize: 16 }} />;
    }
  };

  const getLogChipColor = (level: string) => {
    switch (level) {
      case 'success':
        return 'success';
      case 'error':
        return 'error';
      case 'warning':
        return 'warning';
      default:
        return 'info';
    }
  };

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" p={3}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box p={3}>
        <Alert severity="error">
          Failed to load logs: {(error as any).message}
        </Alert>
      </Box>
    );
  }

  if (!logs || logs.length === 0) {
    return (
      <Box p={3}>
        <Alert severity="info">
          No logs available for this environment
        </Alert>
      </Box>
    );
  }

  return (
    <TableContainer sx={{ maxHeight: 400 }}>
      <Table stickyHeader size="small">
        <TableHead>
          <TableRow>
            <TableCell>Time</TableCell>
            <TableCell>Type</TableCell>
            <TableCell>Level</TableCell>
            <TableCell>Message</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {logs.map((log) => (
            <TableRow key={log.id}>
              <TableCell sx={{ whiteSpace: 'nowrap' }}>
                {format(new Date(log.timestamp), 'HH:mm:ss')}
              </TableCell>
              <TableCell>
                <Chip
                  label={log.type.replace('_', ' ')}
                  size="small"
                  variant="outlined"
                />
              </TableCell>
              <TableCell>
                <Box display="flex" alignItems="center" gap={0.5}>
                  {getLogIcon(log.level)}
                  <Chip
                    label={log.level}
                    size="small"
                    color={getLogChipColor(log.level)}
                  />
                </Box>
              </TableCell>
              <TableCell>
                <Typography variant="body2">{log.message}</Typography>
                {log.details && (
                  <Typography variant="caption" color="text.secondary" component="div">
                    {typeof log.details === 'object' 
                      ? JSON.stringify(log.details, null, 2)
                      : log.details}
                  </Typography>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};
