import { test, expect } from '@playwright/test';

test.describe('Operations Flow', () => {
  test('Admin can access scheduling and email queue', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Sign In');
    await page.getByLabel('Username').fill('admin');
    await page.getByLabel('Password').fill('Password123!');
    await page.click('button[type="submit"]');
    await page.getByRole('button', { name: 'Scheduling Ops' }).click();
    await expect(page.getByRole('heading', { name: 'Scheduling Ops' })).toBeVisible();
    await page.getByRole('button', { name: 'Email Queue' }).click();
    await expect(page.getByRole('heading', { name: 'Email Queue' })).toBeVisible();
  });
});
