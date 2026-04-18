import { test, expect } from '@playwright/test';

test.describe('Analytics Flow', () => {
  test('Admin can access analytics page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Login');
    await page.fill('input[name="username"]', 'admin1');
    await page.fill('input[name="password"]', 'adminpass');
    await page.click('button[type="submit"]');
    await page.click('text=Analytics');
    await expect(page.locator('text=KPIs')).toBeVisible();
  });
});
