import { createTheme } from '@mui/material/styles';

export const theme = createTheme({
  typography: {
    fontFamily: '"Plus Jakarta Sans", "Segoe UI", sans-serif',
    h3: { fontWeight: 700, letterSpacing: -0.5 },
    h4: { fontWeight: 700, letterSpacing: -0.4 },
    h5: { fontWeight: 700, letterSpacing: -0.2 },
    button: { textTransform: 'none', fontWeight: 600 }
  },
  palette: {
    mode: 'light',
    primary: { main: '#0d6e6e' },
    secondary: { main: '#d97706' },
    success: { main: '#2f855a' },
    background: {
      default: '#eef4f7',
      paper: '#ffffff'
    }
  },
  shape: { borderRadius: 14 },
  components: {
    MuiPaper: {
      styleOverrides: {
        root: {
          boxShadow: '0 12px 30px rgba(5, 42, 57, 0.08)'
        }
      }
    },
    MuiButton: {
      styleOverrides: {
        root: { borderRadius: 10, paddingInline: 18 }
      }
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundImage: 'linear-gradient(90deg, #0d6e6e 0%, #1f7a8c 45%, #2a9d8f 100%)'
        }
      }
    }
  }
});
