import { useState, useMemo } from 'react';
import { useApp, useDocumentTheme, useHostStyles } from '@modelcontextprotocol/ext-apps/react';
import { z } from 'zod';
import { LineChart } from '@mui/x-charts/LineChart';
import { ThemeProvider, createTheme, Alert, Box, Typography } from '@mui/material';
import { getCssVar } from '@gke-mcp/ui/shared/utils/styles';
import {
  type AppTimeSeries,
  type ChartDataPoint,
  type ChartSeries,
  transformApiData,
  formatDate,
  appTimeSeriesSchema,
} from './common/utils';
import { TIMESTAMP_KEY } from './common/const';

const MCP_TOOL = {
  QUERY_TIME_SERIES: 'query_time_series',
} as const;

const timeSeriesChartArgsSchema = z.object({
  project_id: z.string().optional(),
  query: z.string(),
  title: z.string().optional(),
  x_legend: z.string().optional(),
  y_legend: z.string().optional(),
});

const queryTimeSeriesRequestSchema = z.object({
  project_id: z.string().optional(),
  query: z.string(),
});

const queryTimeSeriesResponseSchema = z.object({
  data: z.array(appTimeSeriesSchema).nullable(),
});

function TimeSeriesChart({
  data,
  series,
  xLegend,
  yLegend,
  loading,
}: {
  data: ChartDataPoint[];
  series: ChartSeries[];
  xLegend: string;
  yLegend: string;
  loading: boolean;
}) {
  const chartSeries = series.map(({ dataKey, label }) => ({
    dataKey,
    label,
    showMark: false,
  }));

  return (
    <LineChart
      height={300}
      dataset={data}
      loading={loading}
      xAxis={[
        {
          dataKey: TIMESTAMP_KEY,
          scaleType: 'time',
          valueFormatter: formatDate,
          label: xLegend,
          // labelStyle: { fill: getCssVar('--color-text-primary') },
          // tickLabelStyle: { fill: getCssVar('--color-text-primary') },
          // sx: {
          //   '& .MuiChartsAxis-line': { stroke: getCssVar('--color-text-primary') },
          //   '& .MuiChartsAxis-tick': { stroke: getCssVar('--color-text-primary') },
          // },
        },
      ]}
      yAxis={[
        {
          label: yLegend,
          // labelStyle: { fill: getCssVar('--color-text-primary') },
          // tickLabelStyle: { fill: getCssVar('--color-text-primary') },
          // sx: {
          //   '& .MuiChartsAxis-line': { stroke: getCssVar('--color-text-primary') },
          //   '& .MuiChartsAxis-tick': { stroke: getCssVar('--color-text-primary') },
          // },
        },
      ]}
      series={chartSeries}
      slotProps={{
        legend: {
          position: { vertical: 'bottom', horizontal: 'center' },
          sx: {
            height: '100%',
            maxHeight: '10vh',
            overflow: 'auto',
          },
        },
      }}
    />
  );
}

function App() {
  const [data, setData] = useState<AppTimeSeries[]>([]);
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');
  const [title, setTitle] = useState('Timeseries Data Viewer');
  const [xLegend, setXLegend] = useState('Time');
  const [yLegend, setYLegend] = useState('Value');

  const { app } = useApp({
    appInfo: {
      name: 'Time Series Chart',
      version: '1.0.0',
    },
    capabilities: {},
    onAppCreated: (appInstance) => {
      appInstance.ontoolinput = async (request) => {
        try {
          const parseResult = timeSeriesChartArgsSchema.safeParse(request.arguments);

          if (!parseResult.success) {
            throw new Error(
              `Invalid time series parameters provided in tool input:\n${parseResult.error.message}`
            );
          }

          const args = parseResult.data;

          setQuery(args.query);
          if (args.title) {
            setTitle(args.title);
          }
          if (args.x_legend) {
            setXLegend(args.x_legend);
          }
          if (args.y_legend) {
            setYLegend(args.y_legend);
          }

          const toolArgs = queryTimeSeriesRequestSchema.parse(args);
          const response = await appInstance.callServerTool({
            name: MCP_TOOL.QUERY_TIME_SERIES,
            arguments: toolArgs,
          });

          if (response.isError) {
            const errorText =
              response.content?.[0]?.type === 'text' ? response.content[0].text : 'Unknown Error';
            throw new Error(`Failed to call time series API: ${errorText}`);
          } else {
            const parseResult = queryTimeSeriesResponseSchema.safeParse(response.structuredContent);
            if (!parseResult.success) {
              throw new Error(
                `Invalid structured data from time series API:\n${parseResult.error.message}`
              );
            } else {
              const rawData = parseResult.data.data ?? [];
              setData(rawData);
            }
          }
        } catch (err: unknown) {
          const msg = `${err instanceof Error ? err.message : String(err)}`;
          setErrorMsg(msg);
          appInstance
            .updateModelContext({ content: [{ type: 'text', text: msg }] })
            .catch(console.error);
          setData([]);
        } finally {
          setLoading(false);
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
          text: {
            primary: getCssVar('--color-text-primary'),
            secondary: getCssVar('--color-text-secondary'),
            disabled: getCssVar('--color-text-disabled'),
          },
        },
        typography: {
          fontFamily: getCssVar('--font-sans'),
        },
      }),
    [docTheme]
  );

  const transformedData = useMemo(() => transformApiData(data, query), [data, query]);

  return (
    <ThemeProvider theme={theme}>
      {errorMsg ? (
        <Alert severity="error">{errorMsg}</Alert>
      ) : (
        <Box
          sx={{
            padding: '24px',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <Typography sx={{ textAlign: 'center' }} color="text.primary">
            {title}
          </Typography>
          <TimeSeriesChart
            data={transformedData.data}
            series={transformedData.series}
            loading={loading}
            xLegend={xLegend}
            yLegend={yLegend}
          />
        </Box>
      )}
    </ThemeProvider>
  );
}

export default App;
