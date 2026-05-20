import { Box, Paper, Stack, Typography } from "@mui/material";

type JsonPreviewProps = {
  value: unknown;
  title?: string;
};

export default function JsonPreview({ value, title = "JSON preview" }: JsonPreviewProps) {
  return (
    <Stack spacing={1.5}>
      <Typography variant="subtitle2" color="text.secondary">
        {title}
      </Typography>
      <Paper
        variant="outlined"
        sx={{
          bgcolor: "#0E172F",
          color: "#D8E3FF",
          maxHeight: { xs: 320, md: 520 },
          overflow: "auto",
          borderRadius: 1.25,
          p: 1.5,
        }}
      >
        <Box
          component="pre"
          sx={{
            m: 0,
            fontSize: 12,
            lineHeight: 1.5,
            fontFamily: '"SFMono-Regular", Consolas, "Liberation Mono", monospace',
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
          }}
        >
          {JSON.stringify(value, null, 2)}
        </Box>
      </Paper>
    </Stack>
  );
}
