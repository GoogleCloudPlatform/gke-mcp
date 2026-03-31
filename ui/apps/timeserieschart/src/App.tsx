import { useState, useMemo } from 'react';
import { useApp, useDocumentTheme, useHostStyles } from '@modelcontextprotocol/ext-apps/react';
import type { CallToolResult } from '@modelcontextprotocol/sdk/types.js';
import { z } from 'zod';
import { LineChart } from '@mui/x-charts/LineChart';
import { ThemeProvider, createTheme, Alert, Box, Typography } from '@mui/material';
import { getCssVar } from '@gke-mcp/ui/shared/utils/styles';
import type { DatasetElementType } from '@mui/x-charts/internals';

export const MCP_TOOL = {
  QUERY_TIME_SERIES: 'query_time_series',
} as const;

const TIMESTAMP_KEY = 'timestamp' as const;

const timeSeriesChartArgsSchema = z.object({
  project_id: z.string().optional(),
  query: z.string(),
  start_time: z.string().optional(),
  end_time: z.string().optional(),
  title: z.string().optional(),
  x_legend: z.string().optional(),
  y_legend: z.string().optional(),
});

const appTimeSeriesDataPointSchema = z.object({
  timestamp: z.number().optional(),
  value: z.number().optional(),
});

const appTimeSeriesSchema = z.object({
  label: z.string().optional(),
  points: z.array(appTimeSeriesDataPointSchema).optional(),
});

const queryTimeSeriesResponseSchema = z.object({
  data: z.array(appTimeSeriesSchema),
});

type AppTimeSeries = z.infer<typeof appTimeSeriesSchema>;

type ChartDataPoint = DatasetElementType<number | Date> & {
  [TIMESTAMP_KEY]: Date;
};

function transformGCPData(apiResponse: AppTimeSeries[]) {
  if (!apiResponse || apiResponse.length === 0) {
    return { data: [], seriesKeys: [] };
  }

  const timeMap = new Map<number, Record<string, number>>();
  const lineKeys = new Set<string>();

  apiResponse.forEach((series) => {
    const seriesName = series.label || '';
    lineKeys.add(seriesName);

    if (series.points && Array.isArray(series.points)) {
      series.points.forEach((point) => {
        const timestamp = new Date(point.timestamp || 0).getTime();
        const value = point.value ?? 0;

        if (!timeMap.has(timestamp)) {
          timeMap.set(timestamp, {});
        }
        timeMap.get(timestamp)![seriesName] = value;
      });
    }
  });

  const data: ChartDataPoint[] = Array.from(timeMap.entries())
    .sort(([timestampA], [timestampB]) => timestampA - timestampB)
    .map(([timestamp, point]) => ({ ...point, [TIMESTAMP_KEY]: new Date(timestamp) }));

  return { data, seriesKeys: Array.from(lineKeys) };
}

const MonitoringChart = ({
  data,
  seriesKeys,
  xLegend,
  yLegend,
}: {
  data: ChartDataPoint[];
  seriesKeys: string[];
  xLegend: string;
  yLegend: string;
}) => {
  if (data.length === 0) {
    return <Typography>No data available for the specified time range.</Typography>;
  }

  const series = seriesKeys.map((key) => ({
    dataKey: key,
    label: key,
    showMark: false,
  }));

  return (
    <LineChart
      height={300}
      dataset={data}
      xAxis={[
        {
          dataKey: TIMESTAMP_KEY,
          scaleType: 'time',
          valueFormatter: (value: Date) => {
            const dateStr = value.toLocaleDateString('en-US', {
              month: 'short',
              day: 'numeric',
            });
            return `${dateStr} ${value.toLocaleTimeString()}`;
          },
          label: xLegend,
        },
      ]}
      yAxis={[
        {
          label: yLegend,
        },
      ]}
      series={series}
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
};

function App() {
  const [data, setData] = useState<AppTimeSeries[]>([]);
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
            const msg = `Invalid time series parameters provided in tool input:\n${parseResult.error.message}`;
            setErrorMsg(msg);
            setLoading(false);
            appInstance
              .updateModelContext({ content: [{ type: 'text', text: msg }] })
              .catch(console.error);
            return;
          }

          const args = parseResult.data;

          if (args.title) {
            setTitle(args.title);
          }

          if (args.x_legend) {
            setXLegend(args.x_legend);
          }

          if (args.y_legend) {
            setYLegend(args.y_legend);
          }

          const response = (await appInstance.callServerTool({
            name: MCP_TOOL.QUERY_TIME_SERIES,
            arguments: args,
          })) as CallToolResult;

          if (response.isError) {
            const errorText =
              response.content?.[0]?.type === 'text' ? response.content[0].text : 'Unknown Error';
            throw new Error(errorText);
          } else {
            const parseResult = queryTimeSeriesResponseSchema.safeParse(response.structuredContent);
            if (!parseResult.success) {
              throw new Error(
                `Invalid structured data from time series API:\n${parseResult.error.message}`
              );
            } else {
              setData(parseResult.data.data);
            }
          }
        } catch (err: unknown) {
          const msg = `Failed to call time series API: ${err instanceof Error ? err.message : String(err)}`;
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

  const transformedData = useMemo(() => transformGCPData(data), [data]);

  if (loading) {
    return (
      <Box
        sx={{
          padding: '24px',
          height: 400,
        }}
      >
        <Typography textAlign="center">Loading Time Series Data...</Typography>
      </Box>
    );
  }

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
          <Typography textAlign="center">{title}</Typography>
          <MonitoringChart
            data={transformedData.data}
            seriesKeys={transformedData.seriesKeys}
            xLegend={xLegend}
            yLegend={yLegend}
          />
        </Box>
      )}
    </ThemeProvider>
  );
}

export default App;
