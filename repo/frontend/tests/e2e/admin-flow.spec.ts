import { test, expect } from '@playwright/test';

test.describe('Admin Flow', () => {
  test('Admin can assign roles and view audits', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Sign In');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'Password123!');
    await page.click('button[type="submit"]');
    await page.click('text=Admin');
    await page.click('text=Assign Role');
    await page.fill('input[name="targetUserId"]', '2');
    await page.selectOption('select[name="role"]', 'traveler');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Role assigned')).toBeVisible();
    await page.click('text=Role Audits');
    await expect(page.locator('text=Audit Log')).toBeVisible();
  });
});
