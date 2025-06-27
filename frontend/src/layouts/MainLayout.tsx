import React, { useEffect } from 'react';
import { Outlet, useNavigate } from 'react-router-dom';
import {
  Box,
  AppBar,
  Toolbar,
  Typography,
  IconButton,
  Drawer,
  List,
  ListItemIcon,
  ListItemText,
  Divider,
  useTheme,
  Badge,
  Menu,
  MenuItem,
  Avatar,
  Tooltip,
  ListItemButton,
} from '@mui/material';
import {
  Menu as MenuIcon,
  Dashboard,
  Notifications,
  History,
  Logout,
  People,
} from '@mui/icons-material';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

import { useAppSelector, useAppDispatch } from '@/store';
import { toggleSidebar } from '@/store/slices/uiSlice';
import { setUnreadCount } from '@/store/slices/logsSlice';
import { CreateEnvironmentDialog, EditEnvironmentDialog } from '@/components/environments';
import { ConfirmDialog } from '@/components/common/ConfirmDialog';
import { Logo } from '@/components/common/Logo';
import { NotificationPopover } from '@/components/common/NotificationPopover';
import { useNotifications } from '@/hooks/useNotifications';

const DRAWER_WIDTH = 280;
const DRAWER_WIDTH_COLLAPSED = 64;

export const MainLayout: React.FC = () => {
  const theme = useTheme();
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { sidebarOpen, selectedEnvironment } = useAppSelector(state => state.ui);
  const { unreadCount, lastViewedTimestamp } = useAppSelector(state => state.logs);
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const [notificationAnchorEl, setNotificationAnchorEl] = React.useState<null | HTMLElement>(null);

  // Initialize notifications
  useNotifications();

  // Get user info from localStorage
  const userStr = localStorage.getItem('user');
  const user = userStr ? JSON.parse(userStr) : null;

  // Fetch unread logs count
  const { data: logsCountData } = useQuery({
    queryKey: ['logs-count', lastViewedTimestamp],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (lastViewedTimestamp) {
        params.append('since', lastViewedTimestamp);
      }
      const token = localStorage.getItem('authToken');
      const response = await axios.get(`/api/v1/logs/count?${params.toString()}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      return response.data.data;
    },
    refetchInterval: 30000, // Refresh every 30 seconds
    enabled: !!localStorage.getItem('authToken'), // Only fetch if authenticated
  });

  useEffect(() => {
    if (logsCountData?.count !== undefined) {
      dispatch(setUnreadCount(logsCountData.count));
    }
  }, [logsCountData, dispatch]);

  const handleDrawerToggle = () => {
    dispatch(toggleSidebar());
  };

  const handleMenu = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    localStorage.removeItem('authToken');
    localStorage.removeItem('user');
    navigate('/login');
  };

  const handleNotificationClick = (event: React.MouseEvent<HTMLElement>) => {
    setNotificationAnchorEl(event.currentTarget);
  };

  const handleNotificationClose = () => {
    setNotificationAnchorEl(null);
  };

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh', bgcolor: 'background.default' }}>
      <AppBar
        position="fixed"
        elevation={0}
        sx={{
          zIndex: theme.zIndex.drawer + 1,
          height: 64,
        }}
      >
        <Toolbar sx={{ height: '100%' }}>
          <IconButton
            color="inherit"
            aria-label="open drawer"
            edge="start"
            onClick={handleDrawerToggle}
            sx={{ 
              mr: 2,
              transition: 'transform 0.2s',
              transform: sidebarOpen ? 'rotate(180deg)' : 'rotate(0deg)',
            }}
          >
            <MenuIcon />
          </IconButton>
          
          <Box sx={{ display: 'flex', alignItems: 'center', flexGrow: 1, gap: 2 }}>
            <Logo size={40} />
            <Typography variant="h6" noWrap component="div" fontWeight={600}>
              Application Environment Manager
            </Typography>
          </Box>

          <Tooltip title="Notifications">
            <IconButton 
              color="inherit" 
              onClick={handleNotificationClick}
              sx={{ mr: 1 }}
            >
              <Badge badgeContent={unreadCount} color="error">
                <Notifications />
              </Badge>
            </IconButton>
          </Tooltip>
          
          <Tooltip title="Account">
            <IconButton
              color="inherit"
              onClick={handleMenu}
              aria-label="account of current user"
              aria-controls="menu-appbar"
              aria-haspopup="true"
            >
              <Avatar 
                sx={{ 
                  width: 32, 
                  height: 32,
                  bgcolor: 'primary.main',
                  color: 'primary.contrastText',
                  fontSize: '0.875rem',
                  fontWeight: 600,
                }}
              >
                {user?.username?.charAt(0)?.toUpperCase() || 'U'}
              </Avatar>
            </IconButton>
          </Tooltip>
          <Menu
            id="menu-appbar"
            anchorEl={anchorEl}
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'right',
            }}
            keepMounted
            transformOrigin={{
              vertical: 'top',
              horizontal: 'right',
            }}
            open={Boolean(anchorEl)}
            onClose={handleClose}
            PaperProps={{
              sx: {
                mt: 1.5,
                minWidth: 200,
              }
            }}
          >
            <Box sx={{ px: 2, py: 1.5 }}>
              <Typography variant="subtitle2" fontWeight={600}>
                {user?.username || 'User'}
              </Typography>
            </Box>
            <Divider />
            <MenuItem onClick={handleLogout}>
              <ListItemIcon>
                <Logout fontSize="small" />
              </ListItemIcon>
              Logout
            </MenuItem>
          </Menu>
        </Toolbar>
      </AppBar>

      <Drawer
        variant="permanent"
        sx={{
          width: sidebarOpen ? DRAWER_WIDTH : DRAWER_WIDTH_COLLAPSED,
          flexShrink: 0,
          transition: 'width 0.3s ease',
          '& .MuiDrawer-paper': {
            width: sidebarOpen ? DRAWER_WIDTH : DRAWER_WIDTH_COLLAPSED,
            boxSizing: 'border-box',
            pt: '80px',
            transition: 'width 0.3s ease',
            overflowX: 'hidden',
          },
        }}
      >
        <Box sx={{ overflow: 'auto', px: sidebarOpen ? 1 : 0.5 }}>
          <List>
            <Tooltip title={!sidebarOpen ? "Dashboard" : ""} placement="right">
              <ListItemButton 
                onClick={() => navigate('/dashboard')} 
                selected={location.pathname === '/dashboard'}
                sx={{
                  minHeight: 48,
                  justifyContent: sidebarOpen ? 'initial' : 'center',
                  px: sidebarOpen ? 2 : 1,
                }}
              >
                <ListItemIcon
                  sx={{
                    minWidth: 0,
                    mr: sidebarOpen ? 3 : 'auto',
                    justifyContent: 'center',
                  }}
                >
                  <Dashboard />
                </ListItemIcon>
                {sidebarOpen && (
                  <ListItemText 
                    primary="Dashboard" 
                    primaryTypographyProps={{ fontWeight: 500 }}
                  />
                )}
              </ListItemButton>
            </Tooltip>
            <Tooltip title={!sidebarOpen ? "Logs" : ""} placement="right">
              <ListItemButton 
                onClick={() => navigate('/logs')} 
                selected={location.pathname === '/logs'}
                sx={{
                  minHeight: 48,
                  justifyContent: sidebarOpen ? 'initial' : 'center',
                  px: sidebarOpen ? 2 : 1,
                }}
              >
                <ListItemIcon
                  sx={{
                    minWidth: 0,
                    mr: sidebarOpen ? 3 : 'auto',
                    justifyContent: 'center',
                  }}
                >
                  <Badge badgeContent={unreadCount} color="error">
                    <History />
                  </Badge>
                </ListItemIcon>
                {sidebarOpen && (
                  <ListItemText 
                    primary="Logs" 
                    primaryTypographyProps={{ fontWeight: 500 }}
                  />
                )}
              </ListItemButton>
            </Tooltip>
          </List>
          <Divider sx={{ my: 2, mx: sidebarOpen ? 2 : 0.5 }} />
          <List>
            <Tooltip title={!sidebarOpen ? "Users" : ""} placement="right">
              <ListItemButton 
                onClick={() => navigate('/users')} 
                selected={location.pathname === '/users'}
                sx={{
                  minHeight: 48,
                  justifyContent: sidebarOpen ? 'initial' : 'center',
                  px: sidebarOpen ? 2 : 1,
                }}
              >
                <ListItemIcon
                  sx={{
                    minWidth: 0,
                    mr: sidebarOpen ? 3 : 'auto',
                    justifyContent: 'center',
                  }}
                >
                  <People />
                </ListItemIcon>
                {sidebarOpen && (
                  <ListItemText 
                    primary="Users" 
                    primaryTypographyProps={{ fontWeight: 500 }}
                  />
                )}
              </ListItemButton>
            </Tooltip>
          </List>
        </Box>
        <Box sx={{ flexGrow: 1 }} />
        <Box sx={{ px: sidebarOpen ? 1 : 0.5, pb: 2 }}>
          <Divider sx={{ mb: 2, mx: sidebarOpen ? 2 : 0.5 }} />
          <Tooltip title={!sidebarOpen ? "Logout" : ""} placement="right">
            <ListItemButton 
              onClick={handleLogout}
              sx={{
                minHeight: 48,
                justifyContent: sidebarOpen ? 'initial' : 'center',
                px: sidebarOpen ? 2 : 1,
                color: 'error.main',
                '&:hover': {
                  backgroundColor: 'rgba(255, 68, 68, 0.08)',
                },
              }}
            >
              <ListItemIcon
                sx={{
                  minWidth: 0,
                  mr: sidebarOpen ? 3 : 'auto',
                  justifyContent: 'center',
                  color: 'inherit',
                }}
              >
                <Logout />
              </ListItemIcon>
              {sidebarOpen && (
                <ListItemText 
                  primary="Logout" 
                  primaryTypographyProps={{ fontWeight: 500 }}
                />
              )}
            </ListItemButton>
          </Tooltip>
        </Box>
      </Drawer>

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 4,
          mt: '64px',
          minHeight: 'calc(100vh - 64px)',
          backgroundColor: 'background.default',
        }}
      >
        <Box
          sx={{
            maxWidth: 1600,
            margin: '0 auto',
          }}
        >
          <Outlet />
        </Box>
      </Box>

      {/* Dialogs */}
      <CreateEnvironmentDialog />
      <EditEnvironmentDialog environment={selectedEnvironment} />
      <ConfirmDialog />
      
      {/* Notification Popover */}
      <NotificationPopover
        anchorEl={notificationAnchorEl}
        open={Boolean(notificationAnchorEl)}
        onClose={handleNotificationClose}
      />
    </Box>
  );
};
