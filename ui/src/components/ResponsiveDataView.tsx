import {
  Card,
  CardContent,
  Divider,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  useMediaQuery,
} from "@mui/material";
import { useTheme } from "@mui/material/styles";
import type { ReactNode } from "react";

export type ResponsiveColumn<T> = {
  key: string;
  label: string;
  render: (item: T) => ReactNode;
  mobilePriority?: boolean;
  align?: "left" | "right" | "center";
};

type ResponsiveDataViewProps<T> = {
  rows: T[];
  columns: ResponsiveColumn<T>[];
  getRowId: (item: T) => string;
  renderMobileTitle: (item: T) => ReactNode;
  renderMobileSubtitle?: (item: T) => ReactNode;
};

export default function ResponsiveDataView<T>({
  rows,
  columns,
  getRowId,
  renderMobileTitle,
  renderMobileSubtitle,
}: ResponsiveDataViewProps<T>) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down("md"));

  if (isMobile) {
    return (
      <Stack spacing={1.25}>
        {rows.map((row) => (
          <Card key={getRowId(row)} variant="outlined">
            <CardContent sx={{ p: 1.5 }}>
              <Stack spacing={1.25}>
                <Stack spacing={0.5}>
                  <Typography variant="h6" sx={{ fontSize: 14 }}>
                    {renderMobileTitle(row)}
                  </Typography>
                  {renderMobileSubtitle && (
                    <Typography variant="body2" color="text.secondary">
                      {renderMobileSubtitle(row)}
                    </Typography>
                  )}
                </Stack>
                <Divider />
                <Stack spacing={0.9}>
                  {columns
                    .filter((column) => column.mobilePriority !== false)
                    .map((column) => (
                      <Stack key={column.key} direction="row" justifyContent="space-between" spacing={2}>
                        <Typography variant="body2" color="text.secondary" sx={{ fontSize: 12 }}>
                          {column.label}
                        </Typography>
                        <Typography
                          variant="body2"
                          sx={{ textAlign: "right", minWidth: 0, wordBreak: "break-word", fontSize: 12.5 }}
                        >
                          {column.render(row)}
                        </Typography>
                      </Stack>
                    ))}
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        ))}
      </Stack>
    );
  }

  return (
    <TableContainer sx={{ overflowX: "auto" }}>
      <Table size="small" sx={{ minWidth: 860 }}>
        <TableHead>
          <TableRow>
            {columns.map((column) => (
              <TableCell key={column.key} align={column.align ?? "left"}>
                {column.label}
              </TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>
          {rows.map((row) => (
            <TableRow hover key={getRowId(row)}>
              {columns.map((column) => (
                <TableCell key={column.key} align={column.align ?? "left"}>
                  {column.render(row)}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
