import React from 'react';
import { Box } from '@mui/material';

interface LogoProps {
  size?: number;
  color?: string;
}

export const Logo: React.FC<LogoProps> = ({ size = 40, color = '#1976d2' }) => {
  return (
    <Box
      sx={{
        width: size,
        height: size,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <img
        src="/logo.svg"
        alt="Environment Manager Logo"
        style={{
          width: size,
          height: size,
          filter: color !== '#1976d2' ? `brightness(0) saturate(100%) invert(27%) sepia(51%) saturate(2878%) hue-rotate(192deg) brightness(104%) contrast(97%)` : undefined,
        }}
      />
    </Box>
  );
};

export const LogoWithText: React.FC<{ size?: number }> = ({ size = 40 }) => {
  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
      <Logo size={size} />
      <Box>
        <Box sx={{ fontSize: '1.25rem', fontWeight: 700, lineHeight: 1.2 }}>
          Environment
        </Box>
        <Box sx={{ fontSize: '0.875rem', fontWeight: 500, opacity: 0.8 }}>
          Manager
        </Box>
      </Box>
    </Box>
  );
};
