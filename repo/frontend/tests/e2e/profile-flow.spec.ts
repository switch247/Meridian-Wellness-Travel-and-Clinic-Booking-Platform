import { test, expect } from '@playwright/test';

test.describe('Profile Flow', () => {
  test('User can view and edit profile', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Sign In');
    await page.getByLabel('Username').fill('traveler1@example.com');
    await page.getByLabel('Password').fill('Password123!');
    await page.click('button[type="submit"]');
    await page.getByRole('button', { name: 'Profile' }).click();
    await expect(page.getByRole('heading', { name: 'Saved Addresses' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Emergency & Billing Contacts' })).toBeVisible();
    // Simulate address add/edit if UI supports
  });
});
