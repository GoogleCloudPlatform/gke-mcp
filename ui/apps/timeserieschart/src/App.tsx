import { useState, useMemo } from 'react';
import { useApp, useDocumentTheme, useHostStyles } from '@modelcontextprotocol/ext-apps/react';
import type { CallToolResult } from '@modelcontextprotocol/sdk/types.js';
import { z } from 'zod';
import { LineChart } from '@mui/x-charts/LineChart';
import { ThemeProvider, createTheme, Alert, Box } from '@mui/material';
import './App.css';
import { getCssVar } from '@gke-mcp/ui/shared/utils/styles';

export const MCP_TOOL = {
  QUERY_TIME_SERIES: 'query_time_series',
} as const;

const timeSeriesChartArgsSchema = z.object({
  project_id: z.string().optional(),
  query: z.string(),
  start_time: z.string().optional(),
  end_time: z.string().optional(),
  title: z.string().optional(),
  x_legend: z.string().optional(),
  y_legend: z.string().optional(),
});

interface GCPPointData {
  timeInterval?: { endTime?: string };
  values?: Array<{ doubleValue?: number; int64Value?: number }>;
}

interface GCPLabelValue {
  stringValue?: string;
  int64Value?: number;
  boolValue?: boolean;
}

interface GCPSeries {
  labelValues?: GCPLabelValue[];
  pointData?: GCPPointData[];
}

const transformGCPData = (apiResponse: unknown) => {
  if (!apiResponse || !Array.isArray(apiResponse) || apiResponse.length === 0) {
    return { data: [], seriesKeys: [] };
  }

  const timeMap = new Map<number, Record<string, unknown>>();
  const lineKeys = new Set<string>();

  (apiResponse as GCPSeries[]).forEach((series, index) => {
    let seriesName = `series-${index + 1}`;
    if (series.labelValues && Array.isArray(series.labelValues) && series.labelValues.length > 0) {
      seriesName = series.labelValues
        .map((lv) => lv.stringValue ?? lv.int64Value ?? lv.boolValue ?? 'unknown')
        .join('-');
    }

    lineKeys.add(seriesName);

    if (series.pointData && Array.isArray(series.pointData)) {
      series.pointData.forEach((point) => {
        const timestamp = new Date(point.timeInterval?.endTime || 0).getTime();
        const valueObj = point.values?.[0] || {};
        const value = valueObj.doubleValue ?? valueObj.int64Value ?? 0;

        if (!timeMap.has(timestamp)) {
          timeMap.set(timestamp, { timestamp });
        }
        timeMap.get(timestamp)![seriesName] = value;
      });
    }
  });

  const data = Array.from(timeMap.values())
    .sort((a, b) => (a.timestamp as number) - (b.timestamp as number))
    .map((d) => ({ ...d, timestamp: new Date(d.timestamp as number) }));
  return { data, seriesKeys: Array.from(lineKeys) };
};

const MonitoringChart = ({
  rawData,
  xLegend,
  yLegend,
}: {
  rawData: unknown;
  xLegend: string;
  yLegend: string;
}) => {
  const { data, seriesKeys } = transformGCPData(rawData);

  if (data.length === 0) {
    return <p>No data available for the specified time range.</p>;
  }

  const series = seriesKeys.map((key) => ({
    dataKey: key,
    label: key,
    showMark: false,
  }));

  return (
    <LineChart
      dataset={data}
      xAxis={[
        {
          dataKey: 'timestamp',
          scaleType: 'time',
          valueFormatter: (value: unknown) =>
            value instanceof Date
              ? value.toLocaleTimeString()
              : new Date(value as string | number).toLocaleTimeString(),
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
          direction: 'horizontal',
          position: { vertical: 'bottom', horizontal: 'center' },
          classes: {
            root: 'chart-legend-root',
          },
        },
      }}
      sx={{
        '& .MuiChartsAxis-label': {
          fill: getCssVar('--color-text-primary'),
        },
        '& .MuiChartsAxis-tickLabel': {
          fill: getCssVar('--color-text-primary'),
        },
        '& .MuiChartsLegend-label': {
          fill: getCssVar('--color-text-primary'),
        },
      }}
    />
  );
};

function App() {
  const [mcpData, setMcpData] = useState<unknown>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [errorMsg, setErrorMsg] = useState<string>('');
  const [title, setTitle] = useState<string>('Timeseries Data Viewer');
  const [xLegend, setXLegend] = useState<string>('Time');
  const [yLegend, setYLegend] = useState<string>('Value');

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
            const msg = `Error fetching time series: ${errorText}`;
            setErrorMsg(msg);
            appInstance
              .updateModelContext({ content: [{ type: 'text', text: msg }] })
              .catch(console.error);
          } else {
            const contentText =
              response.content?.[0]?.type === 'text' ? response.content[0].text : '[]';
            setMcpData(JSON.parse(contentText));
          }
        } catch (err: unknown) {
          const msg = `Failed to call time series API: ${err instanceof Error ? err.message : String(err)}`;
          setErrorMsg(msg);
          appInstance
            .updateModelContext({ content: [{ type: 'text', text: msg }] })
            .catch(console.error);
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
        },
        typography: {
          fontFamily: getCssVar('--font-sans'),
        },
      }),
    [docTheme]
  );

  if (loading) {
    return (
      <div className="app-wrapper">
        <h2>Loading Time Series Data...</h2>
      </div>
    );
  }

  return (
    <ThemeProvider theme={theme}>
      {errorMsg ? (
        <Alert severity="error">{errorMsg}</Alert>
      ) : (
        <Box
          sx={{
            width: '100%',
            minHeight: 400,
            display: 'flex',
            flex: 1,
            flexDirection: 'column',
            alignItems: 'center',
          }}
        >
          <p>{title}</p>
          <MonitoringChart rawData={mcpData} xLegend={xLegend} yLegend={yLegend} />
        </Box>
      )}
    </ThemeProvider>
  );
}

export default App;
