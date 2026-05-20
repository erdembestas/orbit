import LogoutRoundedIcon from "@mui/icons-material/LogoutRounded";
import MenuRoundedIcon from "@mui/icons-material/MenuRounded";
import {
  AppBar,
  Avatar,
  Box,
  Chip,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Stack,
  Toolbar,
  Typography,
  useMediaQuery,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import { useMemo, useState, type ReactNode } from "react";

export type NavItem = {
  key: string;
  label: string;
  icon: ReactNode;
};

type AppShellProps = {
  title: string;
  username: string;
  currentPage: string;
  navItems: NavItem[];
  onNavigate: (key: string) => void;
  onLogout: () => void;
  children: ReactNode;
};

const drawerWidth = 244;

export default function AppShell({
  title,
  username,
  currentPage,
  navItems,
  onNavigate,
  onLogout,
  children,
}: AppShellProps) {
  const isDesktop = useMediaQuery("(min-width:1024px)");
  const [mobileOpen, setMobileOpen] = useState(false);

  const drawerContent = useMemo(
    () => (
      <Box sx={{ height: "100%", display: "flex", flexDirection: "column", p: 2 }}>
        <Box sx={{ px: 1, pt: 0.5, pb: 2 }}>
          <Typography variant="overline" color="primary.main" sx={{ letterSpacing: "0.16em" }}>
            ORBIT
          </Typography>
          <Typography variant="h6" sx={{ mt: 0.25 }}>
            Control Plane
          </Typography>
          <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, display: "block" }}>
            Phase 1.5 single-cluster mode
          </Typography>
        </Box>

        <Divider sx={{ mb: 1.25 }} />

        <List sx={{ flexGrow: 1, display: "grid", gap: 0.5 }}>
          {navItems.map((item) => {
            const selected = item.key === currentPage;
            return (
              <ListItemButton
                key={item.key}
                selected={selected}
                onClick={() => {
                  onNavigate(item.key);
                  setMobileOpen(false);
                }}
                sx={{
                  minHeight: 40,
                  borderRadius: 1,
                  px: 1.25,
                  borderLeft: "3px solid",
                  borderLeftColor: selected ? "primary.main" : "transparent",
                  bgcolor: selected ? alpha("#1677FF", 0.08) : "transparent",
                  color: selected ? "primary.main" : "text.primary",
                  "&.Mui-selected": {
                    bgcolor: alpha("#1677FF", 0.08),
                  },
                  "&.Mui-selected:hover, &:hover": {
                    bgcolor: alpha("#1677FF", 0.1),
                  },
                }}
              >
                <ListItemIcon
                  sx={{
                    minWidth: 36,
                    color: selected ? "primary.main" : "text.secondary",
                  }}
                >
                  {item.icon}
                </ListItemIcon>
                <ListItemText
                  primary={item.label}
                  primaryTypographyProps={{
                    fontWeight: selected ? 700 : 600,
                    fontSize: 14,
                  }}
                />
              </ListItemButton>
            );
          })}
        </List>

        <Divider sx={{ mt: 1.5, mb: 2 }} />
        <Stack direction="row" spacing={1.25} alignItems="center" sx={{ px: 1 }}>
          <Avatar sx={{ width: 30, height: 30, bgcolor: alpha("#1677FF", 0.12), color: "primary.main", fontSize: 13 }}>
            {username.slice(0, 1).toUpperCase()}
          </Avatar>
          <Box sx={{ minWidth: 0 }}>
            <Typography variant="body2" fontWeight={700} noWrap>
              {username}
            </Typography>
            <Typography variant="caption" color="text.secondary" noWrap>
              Local admin session
            </Typography>
          </Box>
        </Stack>
      </Box>
    ),
    [currentPage, navItems, onNavigate, username],
  );

  return (
    <Box sx={{ minHeight: "100vh", bgcolor: "background.default" }}>
      <AppBar
        position="fixed"
        color="inherit"
        elevation={0}
        sx={{
          width: isDesktop ? `calc(100% - ${drawerWidth}px)` : "100%",
          ml: isDesktop ? `${drawerWidth}px` : 0,
          borderBottom: "1px solid",
          borderColor: "divider",
          bgcolor: alpha("#FFFFFF", 0.96),
          backdropFilter: "blur(12px)",
          boxShadow: "none",
        }}
      >
        <Toolbar sx={{ px: { xs: 2, md: 3 }, minHeight: { xs: 56, md: 56 } }}>
          {!isDesktop && (
            <IconButton
              edge="start"
              color="primary"
              aria-label="open navigation"
              onClick={() => setMobileOpen(true)}
              sx={{ mr: 1 }}
            >
              <MenuRoundedIcon />
            </IconButton>
          )}
          <Box sx={{ flexGrow: 1, minWidth: 0 }}>
            <Typography variant="h6" noWrap>
              {title}
            </Typography>
          </Box>
          <Stack direction="row" spacing={1.5} alignItems="center">
            <Chip label={username} size="small" sx={{ display: { xs: "none", sm: "inline-flex" } }} />
            <IconButton color="primary" aria-label="logout" onClick={onLogout}>
              <LogoutRoundedIcon />
            </IconButton>
          </Stack>
        </Toolbar>
      </AppBar>

      <Box sx={{ display: "flex" }}>
        <Drawer
          variant={isDesktop ? "permanent" : "temporary"}
          open={isDesktop ? true : mobileOpen}
          onClose={() => setMobileOpen(false)}
          ModalProps={{ keepMounted: true }}
          PaperProps={{
            sx: {
              width: drawerWidth,
              borderRight: "1px solid",
              borderColor: "divider",
              bgcolor: "#FFFFFF",
              overflowX: "hidden",
            },
          }}
        >
          {drawerContent}
        </Drawer>

        <Box
          component="main"
          sx={{
            flexGrow: 1,
            width: isDesktop ? `calc(100% - ${drawerWidth}px)` : "100%",
            ml: isDesktop ? `${drawerWidth}px` : 0,
            minWidth: 0,
            px: { xs: 2, sm: 3 },
            pt: { xs: 9, md: 10 },
            pb: 4,
            overflowX: "hidden",
          }}
        >
          {children}
        </Box>
      </Box>
    </Box>
  );
}
