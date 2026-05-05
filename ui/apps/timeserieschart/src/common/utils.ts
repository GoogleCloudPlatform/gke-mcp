import { z } from 'zod';
import type { DatasetElementType } from '@mui/x-charts/internals';
import { TIMESTAMP_KEY } from './const';

/**
 * Zod schema for a single data point in an AppTimeSeries.
 * Consists of a numeric timestamp and a numeric value.
 */
export const appTimeSeriesDataPointSchema = z.object({
  timestamp: z.number(),
  value: z.number(),
});

/**
 * Zod schema for an AppTimeSeries, which contains a label (optional)
 * and an array of points.
 */
export const appTimeSeriesSchema = z.object({
  label: z.string().optional(),
  points: z.array(appTimeSeriesDataPointSchema),
});

export type AppTimeSeries = z.infer<typeof appTimeSeriesSchema>;

/**
 * Represents a single point of data after aggregation for use by MUI Line Charts.
 * Normalizes all series data points onto a shared timestamp.
 */
export type ChartDataPoint = DatasetElementType<number | Date> & {
  [TIMESTAMP_KEY]: Date;
};

export type ChartSeries = {
  dataKey: string;
  label: string;
};

/**
 * Transforms backend API response format into a flattened, aggregated
 * data structure suitable for MUI Line Charts. All series data points are
 * merged and sorted onto accurate shared timestamps.
 *
 * @param apiResponse The list of timeseries objects from the backend.
 * @param originalQuery The fallback query name if a series has no label.
 * @returns Dataset data points array and list of series keys found.
 */
export function transformApiData(apiResponse: AppTimeSeries[], originalQuery: string) {
  if (!apiResponse || apiResponse.length === 0) {
    return { data: [], series: [] as ChartSeries[] };
  }

  const timeMap = new Map<number, Record<string, number>>();
  const series = apiResponse.map((timeSeries, index) => ({
    dataKey: `series_${index}`,
    label: timeSeries.label || originalQuery,
  }));

  apiResponse.forEach((timeSeries, index) => {
    const dataKey = series[index].dataKey;

    if (timeSeries.points && Array.isArray(timeSeries.points)) {
      timeSeries.points.forEach((point) => {
        const timestamp = new Date(point.timestamp).getTime();

        if (!timeMap.has(timestamp)) {
          timeMap.set(timestamp, Object.create(null) as Record<string, number>);
        }

        timeMap.get(timestamp)![dataKey] = point.value;
      });
    }
  });

  const data: ChartDataPoint[] = Array.from(timeMap.entries())
    .sort(([timestampA], [timestampB]) => timestampA - timestampB)
    .map(([timestamp, point]) => Object.assign(Object.create(null), point, { [TIMESTAMP_KEY]: new Date(timestamp) }));

  return { data, series };
}

/**
 * Formats a given Date object into a readable 'Month Day, Hour:Min AM/PM' string.
 * Uses `en-US` locale formatting default.
 *
 * @param value The Date object to format.
 * @returns Localized date-time string.
 */
export function formatDate(value: Date) {
  const dateStr = value.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  });
  return `${dateStr} ${value.toLocaleTimeString()}`;
}
