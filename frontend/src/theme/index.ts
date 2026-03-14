import { createTheme, Theme } from '@mui/material/styles';

// ─── Palette ────────────────────────────────────────────────────────────────
const INDIGO   = '#818cf8'; // primary – electric violet-indigo
const INDIGO_D = '#6366f1'; // darker shade
const INDIGO_L = '#a5b4fc'; // lighter shade

const EMERALD   = '#34d399'; // success
const EMERALD_D = '#10b981';
const CORAL     = '#f87171'; // error
const CORAL_D   = '#ef4444';
const AMBER     = '#fbbf24'; // warning
const AMBER_D   = '#f59e0b';
const SKY       = '#60a5fa'; // info

const BG_BASE  = '#06070a'; // near-black with blue tint
const BG_PAPER = '#0d0e14'; // dark navy card surface
const BG_ELEVATED = '#12141e'; // slightly lifted surface

// ─── Theme ──────────────────────────────────────────────────────────────────
export const darkTheme: Theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: INDIGO,
      light: INDIGO_L,
      dark: INDIGO_D,
      contrastText: '#ffffff',
    },
    secondary: {
      main: EMERALD,
      light: '#6ee7b7',
      dark: EMERALD_D,
      contrastText: '#000000',
    },
    background: {
      default: BG_BASE,
      paper: BG_PAPER,
    },
    text: {
      primary: '#e8eaf0',
      secondary: '#7880a0',
    },
    error: {
      main: CORAL,
      light: '#fca5a5',
      dark: CORAL_D,
    },
    warning: {
      main: AMBER,
      light: '#fde68a',
      dark: AMBER_D,
    },
    success: {
      main: EMERALD,
      light: '#6ee7b7',
      dark: EMERALD_D,
    },
    info: {
      main: SKY,
      light: '#93c5fd',
      dark: '#3b82f6',
    },
    divider: 'rgba(129, 140, 248, 0.08)',
  },

  typography: {
    fontFamily: '"DM Sans", "Helvetica Neue", Arial, sans-serif',
    h1: {
      fontFamily: '"Oxanium", sans-serif',
      fontWeight: 700,
      fontSize: '2.5rem',
      letterSpacing: '-0.02em',
    },
    h2: {
      fontFamily: '"Oxanium", sans-serif',
      fontWeight: 700,
      fontSize: '2rem',
      letterSpacing: '-0.01em',
    },
    h3: {
      fontFamily: '"Oxanium", sans-serif',
      fontWeight: 600,
      fontSize: '1.75rem',
      letterSpacing: '-0.01em',
    },
    h4: {
      fontFamily: '"Oxanium", sans-serif',
      fontWeight: 600,
      fontSize: '1.5rem',
    },
    h5: {
      fontFamily: '"Oxanium", sans-serif',
      fontWeight: 600,
      fontSize: '1.25rem',
    },
    h6: {
      fontFamily: '"Oxanium", sans-serif',
      fontWeight: 600,
      fontSize: '1rem',
    },
    button: {
      textTransform: 'none',
      fontWeight: 600,
      letterSpacing: '0.01em',
    },
    caption: {
      fontFamily: '"DM Sans", sans-serif',
      fontSize: '0.75rem',
      letterSpacing: '0.02em',
    },
    overline: {
      fontFamily: '"Oxanium", sans-serif',
      fontSize: '0.68rem',
      fontWeight: 700,
      letterSpacing: '0.1em',
    },
  },

  shape: {
    borderRadius: 12,
  },

  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          backgroundImage: `
            radial-gradient(ellipse 80% 50% at 20% -10%, rgba(99,102,241,0.07) 0%, transparent 60%),
            radial-gradient(ellipse 60% 40% at 80% 110%, rgba(52,211,153,0.04) 0%, transparent 50%)
          `,
          backgroundAttachment: 'fixed',
          scrollbarColor: '#2a2d3e transparent',
          '&::-webkit-scrollbar, & *::-webkit-scrollbar': {
            width: 8,
            height: 8,
          },
          '&::-webkit-scrollbar-track': {
            background: 'transparent',
          },
          '&::-webkit-scrollbar-thumb, & *::-webkit-scrollbar-thumb': {
            borderRadius: 20,
            border: '2px solid transparent',
            backgroundClip: 'content-box',
            backgroundColor: '#2a2d3e',
            minHeight: 32,
          },
          '&::-webkit-scrollbar-thumb:hover, & *::-webkit-scrollbar-thumb:hover': {
            backgroundColor: '#3d4160',
          },
        },
      },
    },

    MuiCard: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          backgroundColor: BG_PAPER,
          backdropFilter: 'blur(20px)',
          border: '1px solid rgba(129, 140, 248, 0.08)',
          boxShadow: '0 4px 24px rgba(0, 0, 0, 0.4)',
          transition: 'all 0.25s cubic-bezier(0.4, 0, 0.2, 1)',
          '&:hover': {
            transform: 'translateY(-3px)',
            borderColor: 'rgba(129, 140, 248, 0.22)',
            boxShadow: '0 8px 40px rgba(0,0,0,0.55), 0 0 0 1px rgba(129,140,248,0.12)',
          },
        },
      },
    },

    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          backgroundColor: BG_PAPER,
        },
        elevation1: {
          boxShadow: '0 2px 12px rgba(0,0,0,0.35)',
          border: '1px solid rgba(129,140,248,0.07)',
        },
        elevation2: {
          boxShadow: '0 4px 20px rgba(0,0,0,0.45)',
          border: '1px solid rgba(129,140,248,0.09)',
        },
        elevation3: {
          boxShadow: '0 8px 32px rgba(0,0,0,0.55)',
          border: '1px solid rgba(129,140,248,0.12)',
        },
      },
    },

    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 10,
          padding: '9px 22px',
          fontSize: '0.875rem',
          fontWeight: 600,
          transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
        },
        contained: {
          boxShadow: '0 2px 8px rgba(0,0,0,0.3)',
          '&:hover': {
            boxShadow: '0 4px 16px rgba(99,102,241,0.35)',
            transform: 'translateY(-1px)',
          },
          '&:active': { transform: 'translateY(0)' },
        },
        containedPrimary: {
          background: `linear-gradient(135deg, ${INDIGO} 0%, ${INDIGO_D} 100%)`,
          '&:hover': {
            background: `linear-gradient(135deg, ${INDIGO_L} 0%, ${INDIGO} 100%)`,
          },
        },
        outlined: {
          borderWidth: '1.5px',
          '&:hover': {
            borderWidth: '1.5px',
            backgroundColor: 'rgba(129,140,248,0.06)',
          },
        },
      },
    },

    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            backgroundColor: 'rgba(13,14,20,0.6)',
            '& fieldset': {
              borderColor: 'rgba(129,140,248,0.15)',
              borderWidth: '1.5px',
            },
            '&:hover fieldset': {
              borderColor: 'rgba(129,140,248,0.3)',
            },
            '&.Mui-focused fieldset': {
              borderColor: INDIGO,
              borderWidth: '1.5px',
              boxShadow: `0 0 0 3px rgba(99,102,241,0.12)`,
            },
          },
        },
      },
    },

    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 6,
          fontWeight: 600,
          fontSize: '0.72rem',
          letterSpacing: '0.02em',
          backgroundColor: 'rgba(129,140,248,0.1)',
          color: 'rgba(255,255,255,0.85)',
          backdropFilter: 'blur(8px)',
        },
        // Force white text on all semantic color chips
        colorPrimary:   { color: '#ffffff' },
        colorSecondary: { color: '#ffffff' },
        colorSuccess:   { color: '#ffffff' },
        colorWarning:   { color: '#ffffff' },
        colorError:     { color: '#ffffff' },
        colorInfo:      { color: '#ffffff' },
      },
    },

    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: 'rgba(6,7,10,0.85)',
          backdropFilter: 'blur(24px)',
          backgroundImage: 'none',
          borderBottom: '1px solid rgba(129,140,248,0.08)',
          boxShadow: '0 1px 16px rgba(0,0,0,0.3)',
        },
      },
    },

    MuiDrawer: {
      styleOverrides: {
        paper: {
          backgroundColor: 'rgba(8,9,14,0.97)',
          backdropFilter: 'blur(24px)',
          borderRight: '1px solid rgba(129,140,248,0.08)',
          boxShadow: '4px 0 24px rgba(0,0,0,0.4)',
        },
      },
    },

    MuiDivider: {
      styleOverrides: {
        root: {
          borderColor: 'rgba(129,140,248,0.08)',
        },
      },
    },

    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          backgroundColor: BG_ELEVATED,
          backdropFilter: 'blur(12px)',
          border: '1px solid rgba(129,140,248,0.15)',
          fontSize: '0.8rem',
          fontWeight: 500,
          padding: '7px 11px',
          fontFamily: '"DM Sans", sans-serif',
        },
      },
    },

    MuiIconButton: {
      styleOverrides: {
        root: {
          transition: 'all 0.2s ease',
          '&:hover': {
            backgroundColor: 'rgba(129,140,248,0.1)',
            transform: 'scale(1.08)',
          },
        },
      },
    },

    MuiListItemButton: {
      styleOverrides: {
        root: {
          borderRadius: 10,
          margin: '2px 8px',
          transition: 'all 0.2s ease',
          '&:hover': {
            backgroundColor: 'rgba(129,140,248,0.08)',
          },
          '&.Mui-selected': {
            backgroundColor: 'rgba(129,140,248,0.14)',
            borderLeft: `3px solid ${INDIGO}`,
            paddingLeft: '13px',
            '&:hover': {
              backgroundColor: 'rgba(129,140,248,0.18)',
            },
          },
        },
      },
    },

    MuiBadge: {
      styleOverrides: {
        badge: {
          fontWeight: 700,
          fontSize: '0.68rem',
          padding: '0 5px',
          minWidth: '18px',
          height: '18px',
          fontFamily: '"Oxanium", sans-serif',
        },
      },
    },

    MuiLinearProgress: {
      styleOverrides: {
        root: {
          height: 4,
          borderRadius: 2,
          backgroundColor: 'rgba(129,140,248,0.08)',
        },
        bar: {
          borderRadius: 2,
          background: `linear-gradient(90deg, ${INDIGO} 0%, ${INDIGO_D} 100%)`,
        },
      },
    },

    MuiTableHead: {
      styleOverrides: {
        root: {
          '& .MuiTableCell-head': {
            fontFamily: '"Oxanium", sans-serif',
            fontWeight: 600,
            fontSize: '0.78rem',
            letterSpacing: '0.06em',
            textTransform: 'uppercase',
            color: '#7880a0',
            backgroundColor: 'rgba(13,14,20,0.8)',
            borderBottom: '1px solid rgba(129,140,248,0.1)',
            padding: '14px 16px',
          },
        },
      },
    },

    MuiTableRow: {
      styleOverrides: {
        root: {
          transition: 'background-color 0.15s ease',
          '&:hover': {
            backgroundColor: 'rgba(129,140,248,0.04)',
          },
          '& .MuiTableCell-body': {
            borderBottom: '1px solid rgba(129,140,248,0.05)',
            padding: '12px 16px',
            fontSize: '0.875rem',
          },
        },
      },
    },

    MuiTab: {
      styleOverrides: {
        root: {
          fontFamily: '"Oxanium", sans-serif',
          fontWeight: 600,
          fontSize: '0.82rem',
          letterSpacing: '0.04em',
          textTransform: 'none',
          minHeight: 44,
          '&.Mui-selected': {
            color: INDIGO,
          },
        },
      },
    },

    MuiTabs: {
      styleOverrides: {
        indicator: {
          backgroundColor: INDIGO,
          height: 2,
          borderRadius: '2px 2px 0 0',
        },
      },
    },
  },
});
