import React from 'react';
import {
  Box,
  Paper,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  IconButton,
  TextField,
  MenuItem,
  InputAdornment,
  TablePagination,
  CircularProgress,
  Alert,
} from '@mui/material';
import {
  Search,
  Refresh,
  CheckCircle,
  Error,
  Warning,
  Info,
} from '@mui/icons-material';
import { format } from 'date-fns';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

import { useAppDispatch } from '@/store';
import { clearUnreadCount } from '@/store/slices/logsSlice';

interface LogEntry {
  id: string;
  timestamp: string;
  environmentId?: string;
  environmentName?: string;
  userId?: string;
  username?: string;
  type: 'health_check' | 'action' | 'system' | 'error' | 'auth';
  level: 'info' | 'warning' | 'error' | 'success';
  action?: string;
  message: string;
  details?: any;
}

interface LogsResponse {
  logs: LogEntry[];
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
}

const fetchLogs = async (params: {
  page: number;
  limit: number;
  type?: string;
  level?: string;
  search?: string;
}): Promise<LogsResponse> => {
  const queryParams = new URLSearchParams();
  queryParams.append('page', params.page.toString());
  queryParams.append('limit', params.limit.toString());
  
  if (params.type && params.type !== 'all') {
    queryParams.append('type', params.type);
  }
  if (params.level && params.level !== 'all') {
    queryParams.append('level', params.level);
  }
  if (params.search) {
    queryParams.append('search', params.search);
  }

  const response = await axios.get(`/api/v1/logs?${queryParams.toString()}`);
  return response.data.data;
};

export const Logs: React.FC = () => {
  const dispatch = useAppDispatch();
  const [filter, setFilter] = React.useState({
    search: '',
    type: 'all',
    level: 'all',
  });

  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(10);
  const [searchValue, setSearchValue] = React.useState('');

  // Clear unread count when entering logs page
  React.useEffect(() => {
    dispatch(clearUnreadCount());
  }, [dispatch]);

  // Debounce search input
  React.useEffect(() => {
    const timer = setTimeout(() => {
      setFilter(prev => ({ ...prev, search: searchValue }));
    }, 500);

    return () => clearTimeout(timer);
  }, [searchValue]);

  const { data, isLoading, error, refetch } = useQuery<LogsResponse>({
    queryKey: ['logs', page, rowsPerPage, filter],
    queryFn: () => fetchLogs({
      page: page + 1,
      limit: rowsPerPage,
      type: filter.type,
      level: filter.level,
      search: filter.search,
    }),
  });

  const handleRefresh = () => {
    refetch();
  };

  const getLogIcon = (level: string) => {
    switch (level) {
      case 'success':
        return <CheckCircle sx={{ color: 'success.main' }} />;
      case 'error':
        return <Error sx={{ color: 'error.main' }} />;
      case 'warning':
        return <Warning sx={{ color: 'warning.main' }} />;
      default:
        return <Info sx={{ color: 'info.main' }} />;
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

  const handleChangePage = (_event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="50vh">
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

  const logs = data?.logs || [];
  const total = data?.pagination.total || 0;

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" component="h1">
          System Logs
        </Typography>
        <IconButton onClick={handleRefresh}>
          <Refresh />
        </IconButton>
      </Box>

      {/* Filters */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Box display="flex" gap={2} alignItems="center">
          <TextField
            label="Search"
            value={searchValue}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchValue(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <Search />
                </InputAdornment>
              ),
            }}
            sx={{ minWidth: 300 }}
          />
          <TextField
            select
            label="Type"
            value={filter.type}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFilter({ ...filter, type: e.target.value })}
            sx={{ minWidth: 150 }}
          >
            <MenuItem value="all">All Types</MenuItem>
            <MenuItem value="health_check">Health Check</MenuItem>
            <MenuItem value="action">Action</MenuItem>
            <MenuItem value="system">System</MenuItem>
            <MenuItem value="error">Error</MenuItem>
            <MenuItem value="auth">Authentication</MenuItem>
          </TextField>
          <TextField
            select
            label="Level"
            value={filter.level}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFilter({ ...filter, level: e.target.value })}
            sx={{ minWidth: 150 }}
          >
            <MenuItem value="all">All Levels</MenuItem>
            <MenuItem value="info">Info</MenuItem>
            <MenuItem value="success">Success</MenuItem>
            <MenuItem value="warning">Warning</MenuItem>
            <MenuItem value="error">Error</MenuItem>
          </TextField>
        </Box>
      </Paper>

      {/* Logs Table */}
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Timestamp</TableCell>
              <TableCell>Environment</TableCell>
              <TableCell>User</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Level</TableCell>
              <TableCell>Message</TableCell>
              <TableCell>Details</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {logs.map((log: LogEntry) => (
                <TableRow key={log.id}>
                  <TableCell>
                    {format(new Date(log.timestamp), 'PPp')}
                  </TableCell>
                  <TableCell>{log.environmentName || '-'}</TableCell>
                  <TableCell>{log.username || '-'}</TableCell>
                  <TableCell>
                    <Chip
                      label={log.type.replace('_', ' ')}
                      size="small"
                      variant="outlined"
                    />
                  </TableCell>
                  <TableCell>
                    <Box display="flex" alignItems="center" gap={1}>
                      {getLogIcon(log.level)}
                      <Chip
                        label={log.level}
                        size="small"
                        color={getLogChipColor(log.level)}
                      />
                    </Box>
                  </TableCell>
                  <TableCell>{log.message}</TableCell>
                  <TableCell>
                    {log.details && (
                      <Typography variant="caption" component="pre">
                        {JSON.stringify(log.details, null, 2)}
                      </Typography>
                    )}
                  </TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
        <TablePagination
          rowsPerPageOptions={[5, 10, 25, 50]}
          component="div"
          count={total}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </TableContainer>
    </Box>
  );
};
