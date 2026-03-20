import React from 'react';
import {
  Box,
  Paper,
  Typography,
  Chip,
  IconButton,
  TextField,
  InputAdornment,
  CircularProgress,
  Alert,
  Collapse,
  Stack,
  MenuItem,
  TablePagination,
  Tooltip,
} from '@mui/material';
import {
  Search,
  Refresh,
  CheckCircle,
  Error,
  Warning,
  Info,
  KeyboardArrowDown,
  KeyboardArrowUp,
} from '@mui/icons-material';
import { format } from 'date-fns';
import { useQuery, keepPreviousData } from '@tanstack/react-query';
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

const levelColor: Record<string, 'success' | 'error' | 'warning' | 'info' | 'default'> = {
  success: 'success',
  error: 'error',
  warning: 'warning',
  info: 'info',
};

const LEVEL_FILTERS = ['all', 'success', 'error', 'warning', 'info'] as const;
type LevelFilter = (typeof LEVEL_FILTERS)[number];

function LogIcon({ level }: { level: string }) {
  switch (level) {
    case 'success': return <CheckCircle sx={{ color: 'success.main', fontSize: 16 }} />;
    case 'error':   return <Error       sx={{ color: 'error.main',   fontSize: 16 }} />;
    case 'warning': return <Warning     sx={{ color: 'warning.main', fontSize: 16 }} />;
    default:        return <Info        sx={{ color: 'info.main',    fontSize: 16 }} />;
  }
}

// Grid columns: Time | Env | User | Type | Level | Message | expand
const GRID = '160px 140px 120px 130px 110px 1fr auto';

function LogRow({ log }: { log: LogEntry }) {
  const [open, setOpen] = React.useState(false);
  const hasDetails = !!log.details;

  return (
    <>
      <Box
        onClick={() => hasDetails && setOpen((v) => !v)}
        sx={{
          display: 'grid',
          gridTemplateColumns: GRID,
          alignItems: 'center',
          gap: 1.5,
          px: 2,
          py: 1,
          borderBottom: '1px solid',
          borderColor: 'divider',
          cursor: hasDetails ? 'pointer' : 'default',
          '&:hover': hasDetails ? { bgcolor: 'action.hover' } : undefined,
        }}
      >
        {/* Time */}
        <Typography variant="caption" sx={{ fontFamily: '"JetBrains Mono", monospace', color: 'text.secondary', whiteSpace: 'nowrap' }}>
          {format(new Date(log.timestamp), 'MMM d, HH:mm:ss')}
        </Typography>

        {/* Environment */}
        {log.environmentName
          ? <Typography variant="body2" fontWeight={600} sx={{ fontSize: '0.8rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{log.environmentName}</Typography>
          : <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.8rem' }}>—</Typography>
        }

        {/* User */}
        <Typography variant="body2" sx={{ fontSize: '0.8rem', color: log.username ? 'text.primary' : 'text.secondary', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {log.username || '—'}
        </Typography>

        {/* Type */}
        <Chip label={log.type.replace('_', ' ')} size="small" variant="outlined" sx={{ maxWidth: 120 }} />

        {/* Level */}
        <Box display="flex" alignItems="center" gap={0.5}>
          <LogIcon level={log.level} />
          <Chip label={log.level} size="small" color={levelColor[log.level] ?? 'default'} />
        </Box>

        {/* Message */}
        <Typography variant="body2" sx={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {log.message}
        </Typography>

        {/* Expand toggle */}
        {hasDetails ? (
          <IconButton size="small" onClick={(e) => { e.stopPropagation(); setOpen((v) => !v); }}>
            {open ? <KeyboardArrowUp fontSize="small" /> : <KeyboardArrowDown fontSize="small" />}
          </IconButton>
        ) : (
          <Box sx={{ width: 28 }} />
        )}
      </Box>

      {/* Collapsible JSON details */}
      {hasDetails && (
        <Collapse in={open}>
          <Box
            sx={{
              px: 3,
              py: 1.5,
              bgcolor: 'action.selected',
              borderBottom: '1px solid',
              borderColor: 'divider',
            }}
          >
            <Typography
              variant="caption"
              component="pre"
              sx={{
                fontFamily: '"JetBrains Mono", monospace',
                fontSize: '0.72rem',
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
                m: 0,
                color: 'text.secondary',
              }}
            >
              {typeof log.details === 'object'
                ? JSON.stringify(log.details, null, 2)
                : log.details}
            </Typography>
          </Box>
        </Collapse>
      )}
    </>
  );
}

export const Logs: React.FC = () => {
  const dispatch = useAppDispatch();
  const [levelFilter, setLevelFilter] = React.useState<LevelFilter>('all');
  const [typeFilter, setTypeFilter] = React.useState('all');
  const [searchValue, setSearchValue] = React.useState('');
  const [debouncedSearch, setDebouncedSearch] = React.useState('');
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(25);

  React.useEffect(() => { dispatch(clearUnreadCount()); }, [dispatch]);

  React.useEffect(() => {
    const timer = setTimeout(() => { setDebouncedSearch(searchValue); setPage(0); }, 500);
    return () => clearTimeout(timer);
  }, [searchValue]);

  const searchRef = React.useRef<HTMLInputElement>(null);

  const { data, isLoading, isFetching, error, refetch } = useQuery<LogsResponse>({
    queryKey: ['logs', page, rowsPerPage, levelFilter, typeFilter, debouncedSearch],
    queryFn: () => fetchLogs({
      page: page + 1,
      limit: rowsPerPage,
      level: levelFilter,
      type: typeFilter,
      search: debouncedSearch,
    }),
    placeholderData: keepPreviousData,
  });

  // Restore focus to search input after refetch completes
  const wasFetchingRef = React.useRef(false);
  React.useEffect(() => {
    if (wasFetchingRef.current && !isFetching) {
      const active = document.activeElement;
      if (active === document.body || active === null) {
        searchRef.current?.focus();
      }
    }
    wasFetchingRef.current = isFetching;
  }, [isFetching]);

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

      {/* Toolbar: level chips + type filter + search */}
      <Box sx={{ mb: 2, display: 'flex', flexWrap: 'wrap', gap: 1.5, alignItems: 'center' }}>
        <Stack direction="row" spacing={0.75} flexWrap="wrap">
          {LEVEL_FILTERS.map((lvl) => (
            <Chip
              key={lvl}
              label={lvl === 'all' ? 'All Levels' : lvl}
              size="small"
              color={lvl === 'all' ? 'default' : (levelColor[lvl] ?? 'default')}
              variant={levelFilter === lvl ? 'filled' : 'outlined'}
              onClick={() => { setLevelFilter(lvl); setPage(0); }}
              sx={{ cursor: 'pointer', textTransform: 'capitalize' }}
            />
          ))}
        </Stack>

        <TextField
          select size="small" label="Type"
          value={typeFilter}
          onChange={(e) => { setTypeFilter(e.target.value); setPage(0); }}
          sx={{ minWidth: 150, ml: 1 }}
        >
          <MenuItem value="all">All Types</MenuItem>
          <MenuItem value="health_check">Health Check</MenuItem>
          <MenuItem value="action">Action</MenuItem>
          <MenuItem value="system">System</MenuItem>
          <MenuItem value="error">Error</MenuItem>
          <MenuItem value="auth">Authentication</MenuItem>
        </TextField>

        <TextField
          size="small"
          placeholder="Search messages…"
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          inputRef={searchRef}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <Search fontSize="small" />
              </InputAdornment>
            ),
          }}
          sx={{ flex: 1, minWidth: 180 }}
        />
      </Box>

      {/* Log list */}
      <Paper elevation={1} sx={{ position: 'relative', opacity: isFetching ? 0.6 : 1, transition: 'opacity 0.15s' }}>
        {/* Column headers */}
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: GRID,
            gap: 1.5,
            px: 2,
            py: 0.75,
            bgcolor: 'action.selected',
            borderBottom: '1px solid',
            borderColor: 'divider',
            borderRadius: '4px 4px 0 0',
          }}
        >
          {['TIME', 'ENVIRONMENT', 'USER', 'TYPE', 'LEVEL', 'MESSAGE', ''].map((col) => (
            <Typography key={col} variant="caption" sx={{ fontWeight: 700, letterSpacing: '0.06em', color: 'text.secondary' }}>
              {col}
            </Typography>
          ))}
        </Box>

        {/* Rows */}
        <Box>
          {logs.length === 0 ? (
            <Box p={3}>
              <Alert severity="info">No logs match the current filters</Alert>
            </Box>
          ) : (
            logs.map((log: LogEntry) => <LogRow key={log.id} log={log} />)
          )}
        </Box>

        {/* Pagination */}
        <TablePagination
          rowsPerPageOptions={[10, 25, 50, 100]}
          component="div"
          count={total}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={(_, p) => setPage(p)}
          onRowsPerPageChange={(e) => { setRowsPerPage(parseInt(e.target.value, 10)); setPage(0); }}
          sx={{ borderTop: '1px solid', borderColor: 'divider' }}
        />
      </Paper>
    </Box>
  );
};
