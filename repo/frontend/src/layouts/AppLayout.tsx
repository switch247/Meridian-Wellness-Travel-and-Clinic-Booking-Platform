import DashboardRoundedIcon from '@mui/icons-material/DashboardRounded';
import BookOnlineRoundedIcon from '@mui/icons-material/BookOnlineRounded';
import CalendarMonthRoundedIcon from '@mui/icons-material/CalendarMonthRounded';
import PersonRoundedIcon from '@mui/icons-material/PersonRounded';
import AdminPanelSettingsRoundedIcon from '@mui/icons-material/AdminPanelSettingsRounded';
import DescriptionRoundedIcon from '@mui/icons-material/DescriptionRounded';
import LogoutRoundedIcon from '@mui/icons-material/LogoutRounded';
import EventNoteRoundedIcon from '@mui/icons-material/EventNoteRounded';
import ManageAccountsRoundedIcon from '@mui/icons-material/ManageAccountsRounded';
import AssignmentTurnedInRoundedIcon from '@mui/icons-material/AssignmentTurnedInRounded';
import AnalyticsRoundedIcon from '@mui/icons-material/AnalyticsRounded';
import GroupsRoundedIcon from '@mui/icons-material/GroupsRounded';
import NotificationsRoundedIcon from '@mui/icons-material/NotificationsRounded';
import EmailRoundedIcon from '@mui/icons-material/EmailRounded';
import {
  AppBar,
  Avatar,
  Box,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Stack,
  Toolbar,
  Typography
} from '@mui/material';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { canAccess, navItems } from '../app/roleMatrix';

const drawerWidth = 260;

function resolveIcon(name: string) {
  switch (name) {
    case 'dashboard': return <DashboardRoundedIcon />;
    case 'catalog': return <BookOnlineRoundedIcon />;
    case 'booking': return <CalendarMonthRoundedIcon />;
    case 'profile': return <PersonRoundedIcon />;
    case 'admin': return <AdminPanelSettingsRoundedIcon />;
    case 'docs': return <DescriptionRoundedIcon />;
    case 'reservations': return <EventNoteRoundedIcon />;
    case 'agenda': return <AssignmentTurnedInRoundedIcon />;
    case 'sessions': return <AssignmentTurnedInRoundedIcon />;
    case 'ops': return <ManageAccountsRoundedIcon />;
    case 'audits': return <AnalyticsRoundedIcon />;
    case 'community': return <GroupsRoundedIcon />;
    case 'notifications': return <NotificationsRoundedIcon />;
    case 'analytics': return <AnalyticsRoundedIcon />;
    case 'email': return <EmailRoundedIcon />;
    default: return <DashboardRoundedIcon />;
  }
}

export function AppLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { me, logout } = useAuth();
  const visibleNav = navItems.filter((n) => canAccess(me?.roles, n.roles));

  return (
    <Box sx={{ display: 'flex', minHeight: '100%' }}>
      <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar>
          <Typography variant="h6" sx={{ fontWeight: 700, flexGrow: 1 }}>
            Meridian Wellness Operations
          </Typography>
          <Stack direction="row" spacing={1.5} alignItems="center">
            <Avatar sx={{ bgcolor: 'rgba(255,255,255,0.2)' }}>{me?.username?.[0]?.toUpperCase() || 'U'}</Avatar>
            <Box>
              <Typography variant="body2" sx={{ color: 'white' }}>{me?.username || 'Guest'}</Typography>
              <Typography variant="caption" sx={{ color: 'rgba(255,255,255,0.85)' }}>
                {(me?.roles || []).join(' | ') || 'unauthenticated'}
              </Typography>
            </Box>
            <IconButton color="inherit" onClick={() => { logout(); navigate('/login'); }}>
              <LogoutRoundedIcon />
            </IconButton>
          </Stack>
        </Toolbar>
      </AppBar>

      <Drawer
        variant="permanent"
        sx={{
          width: drawerWidth,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: drawerWidth,
            boxSizing: 'border-box',
            borderRight: '1px solid rgba(13, 110, 110, 0.15)',
            backgroundImage: 'linear-gradient(180deg, #ffffff 0%, #f2fbfa 100%)'
          }
        }}
      >
        <Toolbar />
        <Box sx={{ p: 2 }}>
          <Typography variant="overline" color="text.secondary">Navigation</Typography>
        </Box>
        <List sx={{ px: 1 }}>
          {visibleNav.map((item) => (
            <ListItemButton
              key={item.path}
              selected={location.pathname === item.path}
              onClick={() => navigate(item.path)}
              sx={{ mb: 0.5, borderRadius: 2 }}
            >
              <ListItemIcon sx={{ minWidth: 38 }}>{resolveIcon(item.icon)}</ListItemIcon>
              <ListItemText primary={item.label} />
            </ListItemButton>
          ))}
        </List>
        <Divider sx={{ mt: 'auto' }} />
        <Box sx={{ p: 2 }}>
          <Typography variant="caption" color="text.secondary">
            TLS + IP allowlist enforced
          </Typography>
        </Box>
      </Drawer>

      <Box component="main" sx={{ flexGrow: 1, p: 3, mt: 8 }}>
        <Outlet />
      </Box>
    </Box>
  );
}
