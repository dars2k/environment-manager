import React, { useState, useEffect } from 'react';
import {
  Popover,
  Box,
  Typography,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Button,
  Divider,
  CircularProgress,
  IconButton,
} from '@mui/material';
import {
  Info,
  Warning,
  Error as ErrorIcon,
  CheckCircle,
  Close,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';
import { format } from 'date-fns';
import { useAppDispatch, useAppSelector } from '@/store';
import { clearUnreadCount } from '@/store/slices/logsSlice';

interface NotificationPopoverProps {
  anchorEl: HTMLElement | null;
  open: boolean;
  onClose: () => void;
}

interface Log {
  id: string;
  timestamp: string;
  environmentName?: string;
  username?: string;
  type: string;
  level: 'info' | 'warning' | 'error' | 'success';
  action?: string;
  message: string;
  details?: Record<string, any>;
}

const getLevelIcon = (level: string) => {
  switch (level) {
    case 'info':
      return <Info fontSize="small" sx={{ color: 'info.main' }} />;
    case 'warning':
      return <Warning fontSize="small" sx={{ color: 'warning.main' }} />;
    case 'error':
      return <ErrorIcon fontSize="small" sx={{ color: 'error.main' }} />;
    case 'success':
      return <CheckCircle fontSize="small" sx={{ color: 'success.main' }} />;
    default:
      return <Info fontSize="small" />;
  }
};

const getLevelColor = (level: string) => {
  // Return subtle background colors suitable for dark theme
  switch (level) {
    case 'info':
      return 'rgba(33, 150, 243, 0.08)'; // Blue with low opacity
    case 'warning':
      return 'rgba(255, 152, 0, 0.08)'; // Orange with low opacity
    case 'error':
      return 'rgba(255, 68, 68, 0.08)'; // Red with low opacity
    case 'success':
      return 'rgba(0, 230, 118, 0.08)'; // Green with low opacity
    default:
      return 'rgba(255, 255, 255, 0.02)'; // Very subtle white
  }
};

export const NotificationPopover: React.FC<NotificationPopoverProps> = ({
  anchorEl,
  open,
  onClose,
}) => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { lastViewedTimestamp } = useAppSelector(state => state.logs);
  const [viewed, setViewed] = useState(false);

  // Fetch last 5 logs (both read and unread)
  const { data, isLoading, error } = useQuery({
    queryKey: ['recent-logs'],
    queryFn: async () => {
      const params = new URLSearchParams({
        limit: '5',
        page: '1',
      });

      const token = localStorage.getItem('authToken');
      const response = await axios.get(`/api/v1/logs?${params.toString()}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      return response.data.data;
    },
    enabled: open && !!localStorage.getItem('authToken'),
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  // Clear unread count when popover is opened
  useEffect(() => {
    if (open) {
      dispatch(clearUnreadCount());
    }
  }, [open, dispatch]);

  // Reset viewed state when popover closes
  useEffect(() => {
    if (!open) {
      setViewed(false);
    } else {
      // Set viewed to true after a delay when opened
      const timer = setTimeout(() => {
        setViewed(true);
      }, 100);
      return () => clearTimeout(timer);
    }
  }, [open]);

  const handleShowAll = () => {
    onClose();
    navigate('/logs');
  };

  const logs: Log[] = data?.logs || [];

  return (
    <Popover
      open={open}
      anchorEl={anchorEl}
      onClose={onClose}
      anchorOrigin={{
        vertical: 'bottom',
        horizontal: 'right',
      }}
      transformOrigin={{
        vertical: 'top',
        horizontal: 'right',
      }}
      PaperProps={{
        sx: {
          mt: 1,
          width: 400,
          maxHeight: 500,
          overflow: 'hidden',
          overflowX: 'hidden',
          display: 'flex',
          flexDirection: 'column',
        },
      }}
    >
      <Box
        sx={{
          p: 2,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          borderBottom: 1,
          borderColor: 'divider',
        }}
      >
        <Typography variant="h6" fontWeight={600}>
          Notifications
        </Typography>
        <IconButton
          size="small"
          onClick={onClose}
          sx={{ ml: 1 }}
        >
          <Close fontSize="small" />
        </IconButton>
      </Box>

      <Box sx={{ flexGrow: 1, overflow: 'auto', overflowX: 'hidden' }}>
        {isLoading ? (
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              p: 4,
            }}
          >
            <CircularProgress size={40} />
          </Box>
        ) : error ? (
          <Box sx={{ p: 3, textAlign: 'center' }}>
            <Typography color="error">
              Failed to load notifications
            </Typography>
          </Box>
        ) : logs.length === 0 ? (
          <Box sx={{ p: 3, textAlign: 'center' }}>
            <Typography color="text.secondary">
              No new notifications
            </Typography>
          </Box>
        ) : (
          <List disablePadding>
            {logs.map((log, index) => {
              const isUnread = !viewed && lastViewedTimestamp && 
                new Date(log.timestamp) > new Date(lastViewedTimestamp);
              
              return (
                <React.Fragment key={log.id}>
                  <ListItem
                    sx={{
                      px: 2,
                      py: 1.5,
                      backgroundColor: isUnread ? getLevelColor(log.level) : 'transparent',
                      transition: 'background-color 0.3s ease',
                      '&:hover': {
                        backgroundColor: 'action.hover',
                      },
                    }}
                  >
                    <ListItemIcon sx={{ minWidth: 36 }}>
                      {getLevelIcon(log.level)}
                    </ListItemIcon>
                    <ListItemText
                      primary={
                        <Box>
                          <Typography
                            variant="body2"
                            fontWeight={isUnread ? 600 : 400}
                            sx={{ mb: 0.5 }}
                          >
                            {log.message}
                          </Typography>
                          {(log.environmentName || log.username) && (
                            <Typography
                              variant="caption"
                              color="text.secondary"
                              component="div"
                            >
                              {log.environmentName && (
                                <span>Environment: {log.environmentName}</span>
                              )}
                              {log.environmentName && log.username && ' • '}
                              {log.username && (
                                <span>By: {log.username}</span>
                              )}
                            </Typography>
                          )}
                        </Box>
                      }
                      secondary={
                        <Typography
                          variant="caption"
                          color="text.secondary"
                          sx={{ display: 'block', mt: 0.5 }}
                        >
                          {format(new Date(log.timestamp), 'MMM d, yyyy h:mm a')}
                        </Typography>
                      }
                    />
                  </ListItem>
                  {index < logs.length - 1 && <Divider component="li" />}
                </React.Fragment>
              );
            })}
          </List>
        )}
      </Box>

      <Box
        sx={{
          p: 2,
          borderTop: 1,
          borderColor: 'divider',
        }}
      >
        <Button
          fullWidth
          variant="text"
          onClick={handleShowAll}
          sx={{ textTransform: 'none' }}
        >
          Show all notifications
        </Button>
      </Box>
    </Popover>
  );
};
