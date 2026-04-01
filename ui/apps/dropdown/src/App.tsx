import { useState, useMemo, useCallback } from 'react';
import { useApp, useDocumentTheme, useHostStyles } from '@modelcontextprotocol/ext-apps/react';
import { z } from 'zod';
import {
  Box,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  ThemeProvider,
  createTheme,
  Alert,
  type SelectChangeEvent,
} from '@mui/material';

const appInfo = {
  name: 'dropdown-app',
  version: '1.0.0',
};

const ToolInputSchema = z.object({
  options: z.array(z.string()).min(1),
  title: z.string().optional(),
});

const MENU_ITEMS_MAX_HEIGHT = 150;

export default function App() {
  const [options, setOptions] = useState<string[]>([]);
  const [selected, setSelected] = useState('');
  const [submitted, setSubmitted] = useState(false);
  const [title, setTitle] = useState('Select a value');
  const [error, setError] = useState('');

  const { app } = useApp({
    appInfo,
    capabilities: {},
    onAppCreated: (appInstance) => {
      appInstance.ontoolinput = (params) => {
        const parsed = ToolInputSchema.safeParse(params.arguments);
        if (parsed.success) {
          setError('');
          setOptions(parsed.data.options);
          if (parsed.data.title) {
            setTitle(parsed.data.title);
          }
        } else {
          setError(`Invalid tool input: ${parsed.error.message}`);
        }
      };
    },
  });

  useHostStyles(app, app?.getHostContext());
  const docTheme = useDocumentTheme();

  const theme = useMemo(
    () =>
      createTheme({
        palette: {
          mode: docTheme,
        },
        typography: {
          fontFamily: 'var(--font-sans)',
        },
      }),
    [docTheme]
  );

  const handleSelect = useCallback(
    async (e: SelectChangeEvent<string>) => {
      const value = e.target.value;
      setSelected(value);
      setSubmitted(true);

      try {
        await app?.sendMessage({ role: 'user', content: [{ type: 'text', text: value }] });
        await app?.close();
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err));
        setSubmitted(false);
      }
    },
    [app]
  );

  return (
    <ThemeProvider theme={theme}>
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          padding: '24px',
          minHeight: MENU_ITEMS_MAX_HEIGHT + 60,
        }}
      >
        <FormControl fullWidth>
          <InputLabel id="resource-dropdown-label">{title}</InputLabel>
          <Select
            labelId="resource-dropdown-label"
            id="resource-dropdown"
            value={selected}
            label={title}
            onChange={handleSelect}
            disabled={submitted}
            MenuProps={{
              PaperProps: {
                style: {
                  maxHeight: MENU_ITEMS_MAX_HEIGHT,
                },
              },
            }}
          >
            {options.map((opt) => (
              <MenuItem key={opt} value={opt}>
                {opt}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error}
          </Alert>
        )}
      </Box>
    </ThemeProvider>
  );
}
