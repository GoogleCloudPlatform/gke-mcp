import { describe, it, expect } from 'vitest';
import { transformApiData, formatDate, type AppTimeSeries } from './utils';

process.env.TZ = 'UTC';

describe('formatDate', () => {
  it('formats dates consistently in en-US', () => {
    const date = new Date('2026-04-01T12:00:00Z');
    const result = formatDate(date);
    expect(result).toBe('Apr 1 12:00:00 PM');
  });
});

describe('transformApiData', () => {
  it('returns empty lists for empty api response', () => {
    const result = transformApiData([], 'cpu_usage');
    expect(result.data).toEqual([]);
    expect(result.seriesKeys).toEqual([]);
  });

  it('aggregates data points correctly by timestamp and sorts them', () => {
    const apiResponse: AppTimeSeries[] = [
      {
        label: 'web-server-1',
        points: [
          { timestamp: 1711972800000, value: 10 },
          { timestamp: 1711976400000, value: 20 },
        ],
      },
      {
        label: 'web-server-2',
        points: [
          { timestamp: 1711976400000, value: 30 },
          { timestamp: 1711980000000, value: 40 },
        ],
      },
    ];

    const result = transformApiData(apiResponse, 'cpu_usage');

    expect(result.seriesKeys).toEqual(['web-server-1', 'web-server-2']);
    expect(result.data).toEqual([
      { timestamp: new Date(1711972800000), 'web-server-1': 10 },
      { timestamp: new Date(1711976400000), 'web-server-1': 20, 'web-server-2': 30 },
      { timestamp: new Date(1711980000000), 'web-server-2': 40 },
    ]);
  });

  it('uses original query if series label is missing', () => {
    const apiResponse: AppTimeSeries[] = [
      {
        points: [{ timestamp: 1711972800000, value: 10 }],
      },
    ];

    const result = transformApiData(apiResponse, 'compute_query');
    expect(result.seriesKeys).toEqual(['compute_query']);
    expect(result.data).toEqual([{ timestamp: new Date(1711972800000), compute_query: 10 }]);
  });
});
