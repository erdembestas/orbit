import { alpha, createTheme } from "@mui/material/styles";

const colors = {
  background: "#090D11",
  surface: "#11161C",
  elevated: "#151B24",
  border: "#242B36",
  borderSubtle: "#1D2430",
  textPrimary: "#E6EDF3",
  textSecondary: "#9AA4B2",
  textMuted: "#667085",
  primary: "#14B8A6",
  secondary: "#22D3EE",
  warning: "#F59E0B",
  error: "#EF4444",
  high: "#F97316",
  info: "#60A5FA",
  success: "#22C55E",
};

const theme = createTheme({
  palette: {
    mode: "dark",
    primary: {
      main: colors.primary,
      light: "#2DD4BF",
      dark: "#0F9488",
      contrastText: "#041412",
    },
    secondary: {
      main: colors.secondary,
    },
    success: {
      main: colors.success,
    },
    warning: {
      main: colors.warning,
    },
    error: {
      main: colors.error,
    },
    info: {
      main: colors.info,
    },
    text: {
      primary: colors.textPrimary,
      secondary: colors.textSecondary,
    },
    divider: colors.border,
    background: {
      default: colors.background,
      paper: colors.surface,
    },
  },
  shape: {
    borderRadius: 10,
  },
  spacing: 8,
  typography: {
    fontFamily: ['"Inter"', '"Segoe UI"', "system-ui", "sans-serif"].join(","),
    h4: {
      fontWeight: 700,
      fontSize: "1.6rem",
      lineHeight: 1.15,
      letterSpacing: "-0.02em",
    },
    h5: {
      fontWeight: 700,
      fontSize: "1.25rem",
      lineHeight: 1.2,
    },
    h6: {
      fontWeight: 600,
      fontSize: "0.95rem",
      lineHeight: 1.25,
    },
    subtitle1: {
      fontSize: "0.875rem",
      lineHeight: 1.45,
      color: colors.textSecondary,
    },
    body1: {
      fontSize: "0.875rem",
      lineHeight: 1.5,
    },
    body2: {
      fontSize: "0.8125rem",
      lineHeight: 1.5,
    },
    caption: {
      fontSize: "0.75rem",
      lineHeight: 1.45,
      color: colors.textMuted,
    },
    button: {
      textTransform: "none",
      fontWeight: 600,
      fontSize: "0.8125rem",
      letterSpacing: 0,
    },
  },
  shadows: [
    "none",
    "0 6px 18px rgba(0, 0, 0, 0.18)",
    "0 8px 20px rgba(0, 0, 0, 0.18)",
    "0 10px 24px rgba(0, 0, 0, 0.2)",
    ...Array(21).fill("0 10px 24px rgba(0, 0, 0, 0.2)"),
  ] as unknown as typeof createTheme extends (...args: never[]) => infer T ? T["shadows"] : never,
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        ":root": {
          colorScheme: "dark",
        },
        html: {
          backgroundColor: colors.background,
        },
        body: {
          margin: 0,
          backgroundColor: colors.background,
          color: colors.textPrimary,
        },
        "*": {
          boxSizing: "border-box",
        },
        "::-webkit-scrollbar": {
          width: 10,
          height: 10,
        },
        "::-webkit-scrollbar-thumb": {
          backgroundColor: alpha(colors.textMuted, 0.4),
          borderRadius: 999,
          border: `2px solid ${colors.background}`,
        },
        "::-webkit-scrollbar-track": {
          backgroundColor: colors.background,
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 10,
          border: `1px solid ${colors.border}`,
          backgroundImage: "none",
          backgroundColor: colors.surface,
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: 10,
          border: `1px solid ${colors.border}`,
          backgroundColor: colors.surface,
          backgroundImage: "none",
          boxShadow: "0 10px 24px rgba(0, 0, 0, 0.16)",
        },
      },
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          backgroundColor: colors.elevated,
          border: `1px solid ${colors.border}`,
          borderRadius: 12,
          backgroundImage: "none",
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundImage: "none",
        },
      },
    },
    MuiButton: {
      defaultProps: {
        disableElevation: true,
      },
      styleOverrides: {
        root: {
          minHeight: 34,
          borderRadius: 8,
          paddingInline: 14,
          fontSize: "0.8125rem",
          whiteSpace: "nowrap",
        },
        contained: {
          backgroundColor: colors.primary,
          color: "#031311",
          boxShadow: "none",
          "&:hover": {
            backgroundColor: "#10A495",
            boxShadow: "none",
          },
        },
        outlined: {
          borderColor: colors.border,
          color: colors.textPrimary,
          backgroundColor: alpha(colors.surface, 0.6),
          "&:hover": {
            borderColor: alpha(colors.primary, 0.6),
            backgroundColor: alpha(colors.primary, 0.08),
          },
        },
        text: {
          color: colors.textPrimary,
          "&:hover": {
            backgroundColor: alpha(colors.primary, 0.08),
          },
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 999,
          height: 24,
          fontSize: 11,
          fontWeight: 600,
          letterSpacing: 0,
          borderColor: colors.border,
        },
      },
    },
    MuiTableContainer: {
      styleOverrides: {
        root: {
          borderRadius: 10,
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        head: {
          backgroundColor: colors.elevated,
          color: colors.textMuted,
          fontSize: 11,
          textTransform: "uppercase",
          letterSpacing: "0.08em",
          fontWeight: 700,
          borderBottom: `1px solid ${colors.border}`,
          paddingTop: 9,
          paddingBottom: 9,
        },
        body: {
          color: colors.textPrimary,
          borderBottom: `1px solid ${colors.borderSubtle}`,
          paddingTop: 9,
          paddingBottom: 9,
          verticalAlign: "top",
        },
      },
    },
    MuiTextField: {
      defaultProps: {
        size: "small",
      },
    },
    MuiInputLabel: {
      styleOverrides: {
        root: {
          color: colors.textMuted,
        },
      },
    },
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          backgroundColor: colors.elevated,
          color: colors.textPrimary,
          "& fieldset": {
            borderColor: colors.border,
          },
          "&:hover fieldset": {
            borderColor: alpha(colors.primary, 0.55),
          },
          "&.Mui-focused fieldset": {
            borderColor: colors.primary,
            borderWidth: 1,
          },
        },
        input: {
          paddingTop: 10,
          paddingBottom: 10,
        },
      },
    },
    MuiTabs: {
      styleOverrides: {
        indicator: {
          backgroundColor: colors.primary,
          height: 2,
        },
      },
    },
    MuiTab: {
      styleOverrides: {
        root: {
          minHeight: 40,
          paddingInline: 12,
          color: colors.textSecondary,
          fontSize: "0.8125rem",
          "&.Mui-selected": {
            color: colors.textPrimary,
          },
        },
      },
    },
    MuiAlert: {
      styleOverrides: {
        root: {
          borderRadius: 10,
          border: `1px solid ${colors.border}`,
        },
      },
    },
  },
});

export default theme;
