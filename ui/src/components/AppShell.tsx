import LogoutRoundedIcon from "@mui/icons-material/LogoutRounded";
import MenuRoundedIcon from "@mui/icons-material/MenuRounded";
import {
  AppBar,
  Box,
  Chip,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemText,
  Stack,
  Tab,
  Tabs,
  Toolbar,
  Tooltip,
  Typography,
  useMediaQuery,
} from "@mui/material";
import { alpha, useTheme } from "@mui/material/styles";
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

const topbarHeight = 60;

export default function AppShell({
  title,
  username,
  currentPage,
  navItems,
  onNavigate,
  onLogout,
  children,
}: AppShellProps) {
  const theme = useTheme();
  const isDesktop = useMediaQuery(theme.breakpoints.up("lg"));
  const [mobileOpen, setMobileOpen] = useState(false);

  const activeIndex = Math.max(
    navItems.findIndex((item) => item.key === currentPage),
    0,
  );

  const drawerContent = useMemo(
    () => (
      <Box sx={{ height: "100%", display: "flex", flexDirection: "column", bgcolor: "background.paper" }}>
        <Box sx={{ px: 2, py: 2 }}>
          <Stack direction="row" alignItems="center" spacing={1.25}>
            <OrbitMark />
            <Stack direction="row" spacing={1} alignItems="center">
              <Typography sx={{ fontSize: 16, fontWeight: 700 }}>Orbit</Typography>
              <Chip
                size="small"
                label="beta"
                sx={{
                  height: 20,
                  bgcolor: alpha("#14B8A6", 0.12),
                  color: "primary.main",
                  borderColor: alpha("#14B8A6", 0.22),
                }}
                variant="outlined"
              />
            </Stack>
          </Stack>
        </Box>
        <Divider />
        <List sx={{ px: 1.5, py: 1.5, display: "grid", gap: 0.5 }}>
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
                  minHeight: 38,
                  borderRadius: 1.25,
                  border: "1px solid",
                  borderColor: selected ? alpha("#14B8A6", 0.24) : "transparent",
                  bgcolor: selected ? alpha("#14B8A6", 0.08) : "transparent",
                  "&:hover": {
                    bgcolor: alpha("#14B8A6", 0.06),
                  },
                }}
              >
                <ListItemText
                  primary={item.label}
                  primaryTypographyProps={{
                    fontSize: 13,
                    fontWeight: selected ? 700 : 500,
                  }}
                />
              </ListItemButton>
            );
          })}
        </List>
      </Box>
    ),
    [currentPage, navItems, onNavigate],
  );

  return (
    <Box sx={{ minHeight: "100vh", bgcolor: "background.default", color: "text.primary" }}>
      <AppBar
        position="sticky"
        color="transparent"
        elevation={0}
        sx={{
          top: 0,
          zIndex: (muiTheme) => muiTheme.zIndex.drawer + 1,
          borderBottom: "1px solid",
          borderColor: "divider",
          bgcolor: alpha("#090D11", 0.92),
          backdropFilter: "blur(14px)",
        }}
      >
        <Toolbar
          sx={{
            minHeight: `${topbarHeight}px !important`,
            px: { xs: 1.5, sm: 2.5, lg: 3 },
            gap: 2,
          }}
        >
          {!isDesktop && (
            <IconButton
              edge="start"
              color="inherit"
              aria-label="Open navigation"
              onClick={() => setMobileOpen(true)}
              sx={{ border: "1px solid", borderColor: "divider", borderRadius: 1.25 }}
            >
              <MenuRoundedIcon fontSize="small" />
            </IconButton>
          )}

          <Stack direction="row" spacing={1.25} alignItems="center" sx={{ minWidth: 0 }}>
            <OrbitMark />
            <Stack direction="row" spacing={1} alignItems="center" sx={{ minWidth: 0 }}>
              <Typography sx={{ fontSize: 16, fontWeight: 700, lineHeight: 1 }}>
                Orbit
              </Typography>
              <Chip
                size="small"
                label="beta"
                variant="outlined"
                sx={{
                  height: 20,
                  display: { xs: "none", sm: "inline-flex" },
                  bgcolor: alpha("#14B8A6", 0.1),
                  color: "primary.main",
                  borderColor: alpha("#14B8A6", 0.22),
                }}
              />
            </Stack>
          </Stack>

          {isDesktop ? (
            <Tabs
              value={activeIndex}
              onChange={(_, index) => onNavigate(navItems[index].key)}
              variant="scrollable"
              scrollButtons={false}
              sx={{
                minHeight: topbarHeight,
                flex: 1,
                ".MuiTabs-flexContainer": {
                  alignItems: "center",
                  gap: 0.5,
                },
              }}
            >
              {navItems.map((item) => (
                <Tab
                  key={item.key}
                  disableRipple
                  label={item.label}
                  sx={{
                    minHeight: topbarHeight,
                    fontSize: 13,
                    px: 1.5,
                  }}
                />
              ))}
            </Tabs>
          ) : (
            <Typography
              sx={{
                flex: 1,
                fontSize: 13,
                fontWeight: 600,
                color: "text.secondary",
                minWidth: 0,
              }}
              noWrap
            >
              {title}
            </Typography>
          )}

          <Stack direction="row" spacing={1} alignItems="center" sx={{ ml: "auto" }}>
            <Chip
              label="orbit-test"
              variant="outlined"
              sx={{
                display: { xs: "none", md: "inline-flex" },
                bgcolor: alpha("#22D3EE", 0.06),
                color: "text.primary",
                borderColor: "divider",
              }}
            />
            <Chip
              label={username}
              variant="outlined"
              sx={{
                bgcolor: alpha("#FFFFFF", 0.02),
                color: "text.primary",
                borderColor: "divider",
              }}
            />
            <Tooltip title="Logout">
              <IconButton
                aria-label="logout"
                onClick={onLogout}
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 1.25,
                  color: "text.secondary",
                }}
              >
                <LogoutRoundedIcon fontSize="small" />
              </IconButton>
            </Tooltip>
          </Stack>
        </Toolbar>
      </AppBar>

      <Drawer
        open={mobileOpen}
        onClose={() => setMobileOpen(false)}
        ModalProps={{ keepMounted: true }}
        PaperProps={{
          sx: {
            width: 280,
            bgcolor: "background.paper",
            borderRight: "1px solid",
            borderColor: "divider",
            backgroundImage: "none",
          },
        }}
      >
        {drawerContent}
      </Drawer>

      <Box
        component="main"
        sx={{
          px: { xs: 1.5, sm: 2.5, lg: 3 },
          py: 2.5,
          width: "100%",
          overflowX: "hidden",
        }}
      >
        {children}
      </Box>
    </Box>
  );
}

function OrbitMark() {
  return (
    <Box sx={{ position: "relative", width: 24, height: 24, flexShrink: 0 }}>
      <Box
        sx={{
          position: "absolute",
          inset: 0,
          borderRadius: "50%",
          border: "2px solid",
          borderColor: "primary.main",
          opacity: 0.9,
        }}
      />
      <Box
        sx={{
          position: "absolute",
          width: 9,
          height: 9,
          top: 1,
          left: 1,
          borderRadius: "50%",
          bgcolor: "primary.main",
          boxShadow: "0 0 14px rgba(20, 184, 166, 0.4)",
        }}
      />
      <Box
        sx={{
          position: "absolute",
          right: -1,
          bottom: 2,
          width: 9,
          height: 9,
          borderRadius: "50%",
          bgcolor: "#22D3EE",
        }}
      />
    </Box>
  );
}
