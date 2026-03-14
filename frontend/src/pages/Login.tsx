import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  TextField,
  Button,
  Typography,
  Alert,
  CircularProgress,
  InputAdornment,
  IconButton,
} from '@mui/material';
import { Visibility, VisibilityOff, Terminal } from '@mui/icons-material';
import axios from 'axios';

// Animated grid background rendered via inline SVG + CSS keyframes
const GridBackground: React.FC = () => (
  <Box
    aria-hidden
    sx={{
      position: 'fixed',
      inset: 0,
      zIndex: 0,
      overflow: 'hidden',
      pointerEvents: 'none',
    }}
  >
    {/* Grid lines */}
    <svg
      width="100%"
      height="100%"
      xmlns="http://www.w3.org/2000/svg"
      style={{ position: 'absolute', inset: 0 }}
    >
      <defs>
        <pattern id="grid" width="60" height="60" patternUnits="userSpaceOnUse">
          <path
            d="M 60 0 L 0 0 0 60"
            fill="none"
            stroke="rgba(129,140,248,0.06)"
            strokeWidth="1"
          />
        </pattern>
        <radialGradient id="fade" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="transparent" />
          <stop offset="100%" stopColor="#06070a" />
        </radialGradient>
      </defs>
      <rect width="100%" height="100%" fill="url(#grid)" />
      <rect width="100%" height="100%" fill="url(#fade)" />
    </svg>

    {/* Ambient glow blobs */}
    <Box
      sx={{
        position: 'absolute',
        top: '-20%',
        left: '-10%',
        width: '55%',
        height: '55%',
        borderRadius: '50%',
        background: 'radial-gradient(circle, rgba(99,102,241,0.12) 0%, transparent 70%)',
        filter: 'blur(60px)',
        animation: 'float1 8s ease-in-out infinite',
        '@keyframes float1': {
          '0%,100%': { transform: 'translate(0,0)' },
          '50%': { transform: 'translate(30px, 20px)' },
        },
      }}
    />
    <Box
      sx={{
        position: 'absolute',
        bottom: '-15%',
        right: '-5%',
        width: '45%',
        height: '45%',
        borderRadius: '50%',
        background: 'radial-gradient(circle, rgba(52,211,153,0.07) 0%, transparent 70%)',
        filter: 'blur(60px)',
        animation: 'float2 10s ease-in-out infinite',
        '@keyframes float2': {
          '0%,100%': { transform: 'translate(0,0)' },
          '50%': { transform: 'translate(-20px, -15px)' },
        },
      }}
    />
  </Box>
);

export const Login: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [credentials, setCredentials] = useState({ username: '', password: '' });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const response = await axios.post('/api/v1/auth/login', credentials);
      const { token, user } = response.data.data;
      localStorage.setItem('authToken', token);
      localStorage.setItem('user', JSON.stringify(user));
      navigate('/dashboard');
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setCredentials({ ...credentials, [field]: e.target.value });
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
        bgcolor: 'background.default',
      }}
    >
      <GridBackground />

      {/* Login card */}
      <Box
        sx={{
          position: 'relative',
          zIndex: 1,
          width: '100%',
          maxWidth: 420,
          mx: 2,
          animation: 'slideUp 0.5s cubic-bezier(0.4,0,0.2,1)',
          '@keyframes slideUp': {
            from: { opacity: 0, transform: 'translateY(24px)' },
            to:   { opacity: 1, transform: 'translateY(0)' },
          },
        }}
      >
        {/* Card with gradient border */}
        <Box
          sx={{
            p: '1.5px',
            borderRadius: '16px',
            background: 'linear-gradient(135deg, rgba(129,140,248,0.35) 0%, rgba(52,211,153,0.15) 50%, rgba(129,140,248,0.05) 100%)',
          }}
        >
          <Box
            component="form"
            onSubmit={handleSubmit}
            sx={{
              borderRadius: '14.5px',
              bgcolor: 'rgba(13,14,20,0.92)',
              backdropFilter: 'blur(32px)',
              p: 4,
              display: 'flex',
              flexDirection: 'column',
              gap: 2.5,
            }}
          >
            {/* Logo + branding */}
            <Box sx={{ textAlign: 'center', mb: 1 }}>
              <Box
                sx={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: 64,
                  height: 64,
                  borderRadius: '16px',
                  background: 'linear-gradient(135deg, rgba(99,102,241,0.25) 0%, rgba(129,140,248,0.1) 100%)',
                  border: '1px solid rgba(129,140,248,0.2)',
                  mb: 2.5,
                  boxShadow: '0 0 32px rgba(99,102,241,0.2)',
                }}
              >
                <Terminal sx={{ fontSize: 32, color: '#818cf8' }} />
              </Box>
              <Typography
                variant="h4"
                sx={{
                  fontFamily: '"Oxanium", sans-serif',
                  fontWeight: 700,
                  letterSpacing: '-0.01em',
                  mb: 0.5,
                  background: 'linear-gradient(135deg, #e8eaf0 0%, #818cf8 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                }}
              >
                Env Manager
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Sign in to your workspace
              </Typography>
            </Box>

            {error && (
              <Alert
                severity="error"
                sx={{
                  bgcolor: 'rgba(248,113,113,0.08)',
                  border: '1px solid rgba(248,113,113,0.2)',
                  borderRadius: '10px',
                  '& .MuiAlert-icon': { color: '#f87171' },
                }}
              >
                {error}
              </Alert>
            )}

            <TextField
              required
              fullWidth
              id="username"
              label="Username"
              name="username"
              autoComplete="username"
              autoFocus
              value={credentials.username}
              onChange={handleChange('username')}
              disabled={loading}
              size="medium"
            />

            <TextField
              required
              fullWidth
              name="password"
              label="Password"
              type={showPassword ? 'text' : 'password'}
              id="password"
              autoComplete="current-password"
              value={credentials.password}
              onChange={handleChange('password')}
              disabled={loading}
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton
                      onClick={() => setShowPassword(!showPassword)}
                      edge="end"
                      size="small"
                      sx={{ color: 'text.secondary' }}
                    >
                      {showPassword ? <VisibilityOff fontSize="small" /> : <Visibility fontSize="small" />}
                    </IconButton>
                  </InputAdornment>
                ),
              }}
            />

            <Button
              type="submit"
              fullWidth
              variant="contained"
              size="large"
              disabled={loading || !credentials.username || !credentials.password}
              sx={{
                mt: 0.5,
                py: 1.4,
                fontSize: '0.95rem',
                letterSpacing: '0.03em',
                boxShadow: '0 4px 20px rgba(99,102,241,0.4)',
              }}
            >
              {loading ? <CircularProgress size={22} sx={{ color: 'white' }} /> : 'Sign In'}
            </Button>
          </Box>
        </Box>

        {/* Footer hint */}
        <Typography
          variant="caption"
          color="text.secondary"
          align="center"
          display="block"
          sx={{ mt: 2.5, opacity: 0.5 }}
        >
          Application Environment Manager
        </Typography>
      </Box>
    </Box>
  );
};
