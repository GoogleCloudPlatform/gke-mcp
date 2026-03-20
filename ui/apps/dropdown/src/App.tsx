import { useState, useMemo } from "react";
import { useApp, useDocumentTheme, useHostStyles } from "@modelcontextprotocol/ext-apps/react";
import { z } from "zod";
import { FormControl, InputLabel, Select, MenuItem, ThemeProvider, createTheme, type SelectChangeEvent } from "@mui/material";

const appInfo = {
  name: "dropdown-app",
  version: "1.0.0",
};

const ToolInputSchema = z.object({
  options: z.array(z.string()),
  title: z.string().optional(),
});

const MENU_ITEMS_MAX_HEIGHT = 150;

export default function App() {
  const [options, setOptions] = useState<string[]>([]);
  const [selected, setSelected] = useState<string>("");
  const [submitted, setSubmitted] = useState<boolean>(false);
  const [title, setTitle] = useState<string>("Select a value");

  const { app } = useApp({
    appInfo,
    capabilities: {},
    onAppCreated: (app) => {
      app.ontoolinput = (params) => {
        const parsed = ToolInputSchema.safeParse(params.arguments);
        if (parsed.success) {
          setOptions(parsed.data.options);
          if (parsed.data.title) {
            setTitle(parsed.data.title);
          }
        } else {
          console.error("Invalid tool input:", parsed.error);
        }
      };
    },
  });

  useHostStyles(app, app?.getHostContext());
  const docTheme = useDocumentTheme();

  const theme = useMemo(() => createTheme({
    palette: {
      mode: docTheme,
    },
    typography: {
      fontFamily: "var(--font-sans)",
    },
  }), [docTheme]);

  const handleSelect = async (e: SelectChangeEvent<string>) => {
    const value = e.target.value;
    setSelected(value);
    setSubmitted(true);

    try {
      await app?.sendMessage({ role: "user", content: [{ type: "text", text: value }] })
      await app?.close()
    } catch (error) {
      console.error("Failed to send message:", error);
    }

  };

  return (
    <ThemeProvider theme={theme}>
      <div className="container" style={{ minHeight: MENU_ITEMS_MAX_HEIGHT + 60 }}>
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
      </div>
    </ThemeProvider>
  );
}
