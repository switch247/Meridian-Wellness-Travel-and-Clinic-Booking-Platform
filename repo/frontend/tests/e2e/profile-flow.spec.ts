import { test, expect } from '@playwright/test';

test.describe('Profile Flow', () => {
  test('User can view and edit profile', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Sign In');
    await page.fill('input[name="username"]', 'traveler1@example.com');
    await page.fill('input[name="password"]', 'Password123!');
    await page.click('button[type="submit"]');
    await page.click('text=Profile');
    await expect(page.locator('text=Addresses')).toBeVisible();
    await expect(page.locator('text=Contacts')).toBeVisible();
    // Simulate address add/edit if UI supports
  });
});
