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
  Tooltip,
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
  pagination: { page: number; limit: number; total: number };
}

const fetchLogs = async (params: {
  page: number; limit: number; type?: string; level?: string; search?: string;
}): Promise<LogsResponse> => {
  const queryParams = new URLSearchParams();
  queryParams.append('page', params.page.toString());
  queryParams.append('limit', params.limit.toString());
  if (params.type && params.type !== 'all') queryParams.append('type', params.type);
  if (params.level && params.level !== 'all') queryParams.append('level', params.level);
  if (params.search) queryParams.append('search', params.search);
  const response = await axios.get(`/api/v1/logs?${queryParams.toString()}`);
  return response.data.data;
};

// Row tint + left border per log level
const LEVEL_STYLE: Record<string, { bg: string; border: string }> = {
  error:   { bg: 'rgba(248,113,113,0.04)',  border: '#f87171' },
  warning: { bg: 'rgba(251,191,36,0.04)',   border: '#fbbf24' },
  success: { bg: 'rgba(52,211,153,0.03)',   border: '#34d399' },
  info:    { bg: 'rgba(96,165,250,0.03)',   border: '#60a5fa' },
};

const getLevelIcon = (level: string) => {
  const sz = { fontSize: 16 };
  switch (level) {
    case 'success': return <CheckCircle sx={{ ...sz, color: '#34d399' }} />;
    case 'error':   return <Error       sx={{ ...sz, color: '#f87171' }} />;
    case 'warning': return <Warning     sx={{ ...sz, color: '#fbbf24' }} />;
    default:        return <Info        sx={{ ...sz, color: '#60a5fa' }} />;
  }
};

const getLevelChipColor = (level: string): 'success' | 'error' | 'warning' | 'info' => {
  const map: Record<string, 'success' | 'error' | 'warning' | 'info'> = {
    success: 'success', error: 'error', warning: 'warning',
  };
  return map[level] ?? 'info';
};

// Pretty-print JSON details, truncated
const DetailsCell: React.FC<{ details: any }> = ({ details }) => {
  if (!details) return null;
  const str = typeof details === 'string' ? details : JSON.stringify(details);
  const truncated = str.length > 60 ? str.slice(0, 60) + '…' : str;
  return (
    <Tooltip
      title={
        <Box
          component="pre"
          sx={{
            m: 0,
            fontFamily: '"JetBrains Mono", monospace',
            fontSize: '0.72rem',
            maxWidth: 360,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-all',
          }}
        >
          {str}
        </Box>
      }
      placement="left"
    >
      <Box
        component="pre"
        sx={{
          m: 0,
          fontFamily: '"JetBrains Mono", monospace',
          fontSize: '0.68rem',
          color: 'text.secondary',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-all',
          maxWidth: 220,
          cursor: 'help',
        }}
      >
        {truncated}
      </Box>
    </Tooltip>
  );
};

export const Logs: React.FC = () => {
  const dispatch = useAppDispatch();
  const [filter, setFilter] = React.useState({ search: '', type: 'all', level: 'all' });
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(10);
  const [searchValue, setSearchValue] = React.useState('');

  React.useEffect(() => { dispatch(clearUnreadCount()); }, [dispatch]);

  React.useEffect(() => {
    const timer = setTimeout(() => setFilter(prev => ({ ...prev, search: searchValue })), 500);
    return () => clearTimeout(timer);
  }, [searchValue]);

  const { data, isLoading, error, refetch } = useQuery<LogsResponse>({
    queryKey: ['logs', page, rowsPerPage, filter],
    queryFn: () => fetchLogs({ page: page + 1, limit: rowsPerPage, ...filter }),
  });

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
        <Alert severity="error">Failed to load logs: {(error as any).message}</Alert>
      </Box>
    );
  }

  const logs  = data?.logs || [];
  const total = data?.pagination.total || 0;

  return (
    <Box>
      {/* Header */}
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Box>
          <Typography
            variant="h4"
            component="h1"
            sx={{ fontFamily: '"Oxanium", sans-serif', fontWeight: 700 }}
          >
            System Logs
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {total.toLocaleString()} total entries
          </Typography>
        </Box>
        <Tooltip title="Refresh">
          <IconButton onClick={() => refetch()}>
            <Refresh />
          </IconButton>
        </Tooltip>
      </Box>

      {/* Filters */}
      <Paper elevation={1} sx={{ p: 2, mb: 3, display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'center' }}>
        <TextField
          label="Search"
          size="small"
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <Search sx={{ fontSize: 18, color: 'text.secondary' }} />
              </InputAdornment>
            ),
          }}
          sx={{ minWidth: 260, flex: 1 }}
        />
        <TextField
          select size="small" label="Type"
          value={filter.type}
          onChange={(e) => setFilter({ ...filter, type: e.target.value })}
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
          select size="small" label="Level"
          value={filter.level}
          onChange={(e) => setFilter({ ...filter, level: e.target.value })}
          sx={{ minWidth: 140 }}
        >
          <MenuItem value="all">All Levels</MenuItem>
          <MenuItem value="info">Info</MenuItem>
          <MenuItem value="success">Success</MenuItem>
          <MenuItem value="warning">Warning</MenuItem>
          <MenuItem value="error">Error</MenuItem>
        </TextField>
      </Paper>

      {/* Table */}
      <TableContainer component={Paper} elevation={1}>
        <Table size="small">
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
            {logs.map((log: LogEntry) => {
              const ls = LEVEL_STYLE[log.level] ?? LEVEL_STYLE.info;
              return (
                <TableRow
                  key={log.id}
                  sx={{
                    bgcolor: ls.bg,
                    borderLeft: `2px solid ${ls.border}`,
                    '& .MuiTableCell-body': {
                      borderBottom: '1px solid rgba(129,140,248,0.05)',
                    },
                  }}
                >
                  <TableCell sx={{ whiteSpace: 'nowrap', fontFamily: '"JetBrains Mono", monospace', fontSize: '0.72rem' }}>
                    {format(new Date(log.timestamp), 'MMM d, HH:mm')}
                  </TableCell>
                  <TableCell>
                    {log.environmentName
                      ? <Typography variant="body2" fontWeight={600} sx={{ fontSize: '0.8rem' }}>{log.environmentName}</Typography>
                      : <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.8rem' }}>—</Typography>
                    }
                  </TableCell>
                  <TableCell sx={{ fontSize: '0.8rem', color: log.username ? 'text.primary' : 'text.secondary' }}>
                    {log.username || '—'}
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={log.type.replace('_', ' ')}
                      size="small"
                      variant="outlined"
                      sx={{ fontSize: '0.68rem', height: 22 }}
                    />
                  </TableCell>
                  <TableCell>
                    <Box display="flex" alignItems="center" gap={0.75}>
                      {getLevelIcon(log.level)}
                      <Chip
                        label={log.level}
                        size="small"
                        color={getLevelChipColor(log.level)}
                        sx={{ fontSize: '0.68rem', height: 20, fontWeight: 700 }}
                      />
                    </Box>
                  </TableCell>
                  <TableCell sx={{ fontSize: '0.82rem', maxWidth: 260 }}>
                    {log.message}
                  </TableCell>
                  <TableCell>
                    <DetailsCell details={log.details} />
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
        <TablePagination
          rowsPerPageOptions={[10, 25, 50, 100]}
          component="div"
          count={total}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={(_, p) => setPage(p)}
          onRowsPerPageChange={(e) => { setRowsPerPage(parseInt(e.target.value, 10)); setPage(0); }}
          sx={{ borderTop: '1px solid rgba(129,140,248,0.07)' }}
        />
      </TableContainer>
    </Box>
  );
};
