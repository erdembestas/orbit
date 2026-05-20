import { alpha, createTheme } from "@mui/material/styles";

const primaryMain = "#1677FF";
const primaryLight = "#4E97FF";
const borderColor = "#DFE3E8";
const textPrimary = "#1F2933";
const textSecondary = "#5F6B7A";

const theme = createTheme({
  palette: {
    mode: "light",
    primary: {
      main: primaryMain,
      light: primaryLight,
      dark: "#0B5FCC",
      contrastText: "#FFFFFF",
    },
    secondary: {
      main: "#1F2933",
    },
    success: {
      main: "#2E7D32",
    },
    warning: {
      main: "#F59E0B",
    },
    error: {
      main: "#D32F2F",
    },
    info: {
      main: "#2563EB",
    },
    text: {
      primary: textPrimary,
      secondary: textSecondary,
    },
    divider: borderColor,
    background: {
      default: "#F7F8FA",
      paper: "#FFFFFF",
    },
  },
  shape: {
    borderRadius: 8,
  },
  spacing: 8,
  typography: {
    fontFamily: ['"Inter"', '"Segoe UI"', "system-ui", "sans-serif"].join(","),
    h4: {
      fontWeight: 700,
      fontSize: "1.625rem",
      letterSpacing: "-0.02em",
    },
    h6: {
      fontWeight: 700,
      fontSize: "1rem",
    },
    subtitle1: {
      color: textSecondary,
      fontSize: "0.92rem",
    },
    body1: {
      fontSize: "0.875rem",
    },
    body2: {
      fontSize: "0.8125rem",
    },
    button: {
      fontWeight: 600,
      textTransform: "none",
      letterSpacing: 0,
    },
  },
  shadows: [
    "none",
    "0 1px 2px rgba(16, 24, 40, 0.04)",
    "0 2px 6px rgba(16, 24, 40, 0.05)",
    "0 3px 8px rgba(16, 24, 40, 0.06)",
    "0 4px 10px rgba(16, 24, 40, 0.06)",
    "0 4px 12px rgba(16, 24, 40, 0.07)",
    ...Array(19).fill("0 4px 12px rgba(16, 24, 40, 0.07)"),
  ] as unknown as typeof createTheme extends (...args: never[]) => infer T ? T["shadows"] : never,
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          background: "#F7F8FA",
        },
        "*": {
          boxSizing: "border-box",
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          border: `1px solid ${borderColor}`,
          backgroundImage: "none",
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          border: `1px solid ${borderColor}`,
          boxShadow: "0 1px 2px rgba(16, 24, 40, 0.04)",
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
          borderRadius: 7,
          paddingInline: 14,
          minHeight: 34,
          fontSize: "0.8125rem",
        },
        contained: {
          color: "#FFFFFF",
          background: primaryMain,
          boxShadow: "none",
          "&:hover": {
            boxShadow: "none",
            background: "#0B6AD9",
          },
        },
        outlined: {
          borderColor: alpha(primaryMain, 0.28),
          color: primaryMain,
          backgroundColor: "#FFFFFF",
          "&:hover": {
            borderColor: alpha(primaryMain, 0.42),
            backgroundColor: alpha(primaryMain, 0.04),
          },
        },
        text: {
          "&:hover": {
            backgroundColor: alpha(primaryMain, 0.06),
          },
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 999,
          fontWeight: 700,
          height: 24,
          fontSize: 11,
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        head: {
          backgroundColor: "#F8F9FB",
          color: textSecondary,
          fontSize: 11,
          textTransform: "uppercase",
          letterSpacing: "0.06em",
          fontWeight: 700,
          borderBottom: `1px solid ${borderColor}`,
          paddingTop: 10,
          paddingBottom: 10,
        },
        body: {
          borderBottom: `1px solid ${alpha("#D0D5DD", 0.55)}`,
          verticalAlign: "top",
          paddingTop: 10,
          paddingBottom: 10,
        },
      },
    },
    MuiTextField: {
      defaultProps: {
        size: "small",
      },
    },
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          backgroundColor: "#FFFFFF",
          "& fieldset": {
            borderColor: borderColor,
          },
          "&:hover fieldset": {
            borderColor: alpha(primaryMain, 0.4),
          },
          "&.Mui-focused fieldset": {
            borderWidth: 1,
            borderColor: primaryMain,
          },
        },
      },
    },
  },
});

export default theme;
