import React from 'react';
import {
  Box,
  Chip,
  CircularProgress,
  Alert,
  Typography,
  InputAdornment,
  TextField,
  Collapse,
  IconButton,
  Stack,
} from '@mui/material';
import { CheckCircle, Error, Warning, Info, Search, KeyboardArrowDown, KeyboardArrowUp } from '@mui/icons-material';
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

const LEVEL_FILTERS = ['all', 'success', 'error', 'warning', 'info'] as const;
type LevelFilter = (typeof LEVEL_FILTERS)[number];

const levelColor: Record<string, 'success' | 'error' | 'warning' | 'info' | 'default'> = {
  success: 'success',
  error: 'error',
  warning: 'warning',
  info: 'info',
};

function LogIcon({ level }: { level: string }) {
  switch (level) {
    case 'success': return <CheckCircle sx={{ color: 'success.main', fontSize: 16 }} />;
    case 'error':   return <Error       sx={{ color: 'error.main',   fontSize: 16 }} />;
    case 'warning': return <Warning     sx={{ color: 'warning.main', fontSize: 16 }} />;
    default:        return <Info        sx={{ color: 'info.main',    fontSize: 16 }} />;
  }
}

function LogRow({ log }: { log: LogEntry }) {
  const [open, setOpen] = React.useState(false);
  const hasDetails = !!log.details;

  return (
    <>
      <Box
        onClick={() => hasDetails && setOpen((v) => !v)}
        sx={{
          display: 'grid',
          gridTemplateColumns: '160px 130px 110px 1fr auto',
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
        {/* Time — full date + time */}
        <Typography variant="caption" sx={{ fontFamily: '"JetBrains Mono", monospace', color: 'text.secondary', whiteSpace: 'nowrap' }}>
          {format(new Date(log.timestamp), 'MMM d, HH:mm:ss')}
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

export const EnvironmentLogs: React.FC<EnvironmentLogsProps> = ({ environmentId }) => {
  const [levelFilter, setLevelFilter] = React.useState<LevelFilter>('all');
  const [search, setSearch] = React.useState('');

  const { data: logs, isLoading, error } = useQuery({
    queryKey: ['environment-logs', environmentId],
    queryFn: async () => {
      const response = await axios.get(`/api/v1/environments/${environmentId}/logs?limit=50`);
      return response.data.data.logs as LogEntry[];
    },
    refetchInterval: 10000,
  });

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" p={3}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box p={2}>
        <Alert severity="error">Failed to load logs: {(error as any).message}</Alert>
      </Box>
    );
  }

  if (!logs || logs.length === 0) {
    return (
      <Box p={2}>
        <Alert severity="info">No logs available for this environment</Alert>
      </Box>
    );
  }

  const filtered = logs.filter((log) => {
    const matchesLevel = levelFilter === 'all' || log.level === levelFilter;
    const matchesSearch = !search || log.message.toLowerCase().includes(search.toLowerCase());
    return matchesLevel && matchesSearch;
  });

  return (
    <Box>
      {/* Toolbar */}
      <Box sx={{ px: 2, pt: 2, pb: 1.5, display: 'flex', flexWrap: 'wrap', gap: 1.5, alignItems: 'center' }}>
        <Stack direction="row" spacing={0.75} flexWrap="wrap">
          {LEVEL_FILTERS.map((lvl) => (
            <Chip
              key={lvl}
              label={lvl === 'all' ? `All (${logs.length})` : lvl}
              size="small"
              color={lvl === 'all' ? 'default' : (levelColor[lvl] ?? 'default')}
              variant={levelFilter === lvl ? 'filled' : 'outlined'}
              onClick={() => setLevelFilter(lvl)}
              sx={{ cursor: 'pointer', textTransform: 'capitalize' }}
            />
          ))}
        </Stack>
        <TextField
          size="small"
          placeholder="Search messages…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <Search fontSize="small" />
              </InputAdornment>
            ),
          }}
          sx={{ ml: 'auto', width: 220 }}
        />
      </Box>

      {/* Column headers */}
      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: '160px 130px 110px 1fr auto',
          gap: 1.5,
          px: 2,
          py: 0.75,
          bgcolor: 'action.selected',
          borderTop: '1px solid',
          borderBottom: '1px solid',
          borderColor: 'divider',
        }}
      >
        {['TIME', 'TYPE', 'LEVEL', 'MESSAGE', ''].map((col) => (
          <Typography key={col} variant="caption" sx={{ fontWeight: 700, letterSpacing: '0.06em', color: 'text.secondary' }}>
            {col}
          </Typography>
        ))}
      </Box>

      {/* Rows */}
      <Box sx={{ maxHeight: 420, overflowY: 'auto' }}>
        {filtered.length === 0 ? (
          <Box p={3}>
            <Alert severity="info">No logs match the current filters</Alert>
          </Box>
        ) : (
          filtered.map((log) => <LogRow key={log.id} log={log} />)
        )}
      </Box>

      {/* Footer count */}
      <Box sx={{ px: 2, py: 1, borderTop: '1px solid', borderColor: 'divider' }}>
        <Typography variant="caption" color="text.secondary">
          Showing {filtered.length} of {logs.length} log entries
        </Typography>
      </Box>
    </Box>
  );
};
