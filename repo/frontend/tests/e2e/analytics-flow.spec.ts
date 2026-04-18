import { test, expect } from '@playwright/test';

test.describe('Analytics Flow', () => {
  test('Admin can access analytics page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Sign In');
    await page.getByLabel('Username').fill('admin');
    await page.getByLabel('Password').fill('Password123!');
    await page.click('button[type="submit"]');
    await page.getByRole('button', { name: 'Analytics' }).click();
    await expect(page.getByRole('heading', { name: 'Analytics' })).toBeVisible();
  });
});
