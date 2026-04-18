import { test, expect } from '@playwright/test';

test.describe('Booking Flow', () => {
  test('User can complete a booking', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Login');
    await page.fill('input[name="username"]', 'traveler1');
    await page.fill('input[name="password"]', 'password1');
    await page.click('button[type="submit"]');
    await page.click('text=Catalog');
    await page.click('text=Book');
    await page.fill('input[name="packageId"]', '1');
    await page.fill('input[name="hostId"]', '1');
    await page.fill('input[name="roomId"]', '1');
    await page.fill('input[name="slotStart"]', '2026-04-17T10:00');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Booking Confirmed')).toBeVisible();
  });
});
