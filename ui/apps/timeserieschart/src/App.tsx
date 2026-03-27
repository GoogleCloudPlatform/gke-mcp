import { useState } from 'react'
import { useApp } from '@modelcontextprotocol/ext-apps/react'
import type { CallToolResult } from '@modelcontextprotocol/sdk/types.js'
import { z } from 'zod'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import './App.css'

const timeSeriesChartArgsSchema = z.object({
  project_id: z.string().optional(),
  query: z.string(),
  start_time: z.string().optional(),
  end_time: z.string().optional(),
  title: z.string().optional(),
  x_legend: z.string().optional(),
  y_legend: z.string().optional(),
})

const COLORS = [
  '#4285F4', '#EA4335', '#FBBC04', '#34A853',
  '#8F00FF', '#00E5FF', '#FF6D00', '#B00020',
  '#3F51B5', '#E91E63', '#009688', '#FF9800'
];

const transformGCPData = (apiResponse: any) => {
  if (!apiResponse || !Array.isArray(apiResponse) || apiResponse.length === 0) {
    return { data: [], lines: [] };
  }

  const timeMap = new Map<number, any>();
  const lineKeys = new Set<string>();

  apiResponse.forEach((series: any, index: number) => {
    // Generate an unique key for the line based on labelValues, fallback to index
    let seriesName = `series-${index + 1}`;
    if (series.labelValues && Array.isArray(series.labelValues) && series.labelValues.length > 0) {
      seriesName = series.labelValues
        .map((lv: any) => lv.stringValue ?? lv.int64Value ?? lv.boolValue ?? 'unknown')
        .join('-');
    }

    lineKeys.add(seriesName);

    if (series.pointData && Array.isArray(series.pointData)) {
      series.pointData.forEach((point: any) => {
        const timestamp = new Date(point.timeInterval?.endTime || 0).getTime();
        const valueObj = point.values?.[0] || {};
        const value = valueObj.doubleValue ?? valueObj.int64Value ?? 0;

        if (!timeMap.has(timestamp)) {
          timeMap.set(timestamp, { timestamp });
        }
        timeMap.get(timestamp)[seriesName] = value;
      });
    }
  });

  const data = Array.from(timeMap.values()).sort((a, b) => a.timestamp - b.timestamp);
  return { data, lines: Array.from(lineKeys) };
};

const MonitoringChart = ({ rawData, xLegend, yLegend }: { rawData: any, xLegend: string, yLegend: string }) => {
  const { data, lines } = transformGCPData(rawData);

  if (data.length === 0) {
    return <div style={{ padding: '20px', textAlign: 'center' }}>No data available for the specified time range.</div>;
  }

  return (
    <div className="chart-wrapper">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 20, right: 30, left: 30, bottom: 40 }}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} />
          <XAxis
            dataKey="timestamp"
            type="number"
            domain={['dataMin', 'dataMax']}
            tickFormatter={(unixTime) => new Date(unixTime).toLocaleTimeString()}
            label={{ value: xLegend, position: 'insideBottom', offset: -25, fill: '#333' }}
          />
          <YAxis label={{ value: yLegend, angle: -90, position: 'insideLeft', offset: -20, fill: '#333' }} />
          <Tooltip
            labelFormatter={(value) => new Date(value as number).toLocaleString()}
            itemSorter={(item) => -(item.value as number)}
            formatter={(value: any, name: any) => {
              if (typeof value === 'number') {
                return [Number.isInteger(value) ? value : value.toFixed(2), name];
              }
              return [value, name];
            }}
            contentStyle={{ maxHeight: '250px', overflowY: 'auto', fontSize: '12px', padding: '10px' }}
          />
          <Legend
            verticalAlign="bottom"
            wrapperStyle={{
              paddingTop: '40px',
              maxHeight: '100px',
              overflowY: 'auto',
              borderTop: '1px solid #eee'
            }}
          />
          {lines.map((lineKey, index) => (
            <Line
              key={lineKey}
              type="monotone"
              dataKey={lineKey}
              stroke={COLORS[index % COLORS.length]}
              dot={false}
              strokeWidth={2}
              isAnimationActive={false}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
};

function App() {
  const [mcpData, setMcpData] = useState<any>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [errorMsg, setErrorMsg] = useState<string>('')
  const [title, setTitle] = useState<string>('Timeseries Data Viewer')
  const [xLegend, setXLegend] = useState<string>('Time')
  const [yLegend, setYLegend] = useState<string>('Value')

  useApp({
    appInfo: {
      name: 'Time Series Chart',
      version: '1.0.0',
    },
    capabilities: {},
    onAppCreated: (appInstance) => {
      appInstance.ontoolinput = async (request) => {
        try {
          const parseResult = timeSeriesChartArgsSchema.safeParse(request.arguments)
          if (!parseResult.success) {
            const msg = `Invalid time series parameters provided in tool input:\n${parseResult.error.message}`
            setErrorMsg(msg)
            setLoading(false)
            appInstance.updateModelContext({ content: [{ type: "text", text: msg }] }).catch(console.error)
            return
          }

          const args = parseResult.data

          if (args.title) {
            setTitle(args.title)
          }

          if (args.x_legend) {
            setXLegend(args.x_legend)
          }

          if (args.y_legend) {
            setYLegend(args.y_legend)
          }

          const response = await appInstance.callServerTool({
            name: 'query_time_series',
            arguments: args
          }) as CallToolResult

          if (response.isError) {
            const errorText = response.content?.[0]?.type === 'text' ? response.content[0].text : 'Unknown Error';
            const msg = `Error fetching time series: ${errorText}`
            setErrorMsg(msg)
            appInstance.updateModelContext({ content: [{ type: "text", text: msg }] }).catch(console.error)
          } else {
            const contentText = response.content?.[0]?.type === 'text' ? response.content[0].text : '[]';
            setMcpData(JSON.parse(contentText))
          }
        } catch (err: any) {
          const msg = `Failed to call time series API: ${err.message}`
          setErrorMsg(msg)
          appInstance.updateModelContext({ content: [{ type: "text", text: msg }] }).catch(console.error)
        } finally {
          setLoading(false)
        }
      }
    }
  })

  if (loading) {
    return <div className="app-wrapper"><h2>Loading Time Series Data...</h2></div>
  }

  if (errorMsg) {
    return <div className="app-wrapper error-message"><h2>Error</h2><p>{errorMsg}</p></div>
  }

  return (
    <div className="app-wrapper">
      <h2 className="chart-title">{title}</h2>
      <MonitoringChart rawData={mcpData} xLegend={xLegend} yLegend={yLegend} />
    </div>
  )
}

export default App
