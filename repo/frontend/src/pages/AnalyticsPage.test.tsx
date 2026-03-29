import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { AnalyticsPage } from './AnalyticsPage';

const mocks = vi.hoisted(() => ({
  scheduleReport: vi.fn().mockResolvedValue({ id: 12 }),
  analyticsKpis: vi.fn().mockResolvedValue({ kpis: {} }),
  exportAnalytics: vi.fn().mockResolvedValue({ path: '/tmp/a.csv' })
}));

vi.mock('../context/AuthContext', () => ({
  useAuth: () => ({ token: 't' })
}));

vi.mock('../api/client', () => ({
  api: {
    analyticsKpis: mocks.analyticsKpis,
    exportAnalytics: mocks.exportAnalytics,
    scheduleReport: mocks.scheduleReport
  }
}));

describe('AnalyticsPage', () => {
  it('submits schedule report action', async () => {
    const u = userEvent.setup();
    render(<AnalyticsPage />);
    const btn = await screen.findByRole('button', { name: /schedule report/i });
    await u.click(btn);
    expect(mocks.scheduleReport).toHaveBeenCalled();
  });
});
