import { test, expect } from '@playwright/test';

test.describe('Operations Flow', () => {
  test('Admin can access scheduling and email queue', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'Password123!');
    await page.click('button[type="submit"]');
    await page.click('text=Operations');
    await page.click('text=Scheduling');
    await expect(page.locator('text=Ops Scheduling')).toBeVisible();
    await page.click('text=Email Queue');
    await expect(page.locator('text=Email Queue')).toBeVisible();
  });
});
