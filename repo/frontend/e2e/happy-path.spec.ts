import { expect, test } from '@playwright/test';

test('traveler happy path: login, catalog, community, analytics schedule guarded', async ({ page }) => {
  await page.goto('/login');

  const username = `e2e_${Date.now()}`;
  await page.getByLabel('Username').fill(username);
  await page.getByLabel('Password').fill('Strong#Pass123');
  await page.getByRole('button', { name: 'Quick Register + Sign In' }).click();

  await expect(page.getByText('Operational Dashboard')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'My Reservations' })).toBeVisible();

  await page.getByRole('button', { name: 'Catalog' }).click();
  await expect(page.getByText('Travel Catalog')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Routes' })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Partner Hotels' })).toBeVisible();

  await page.getByRole('button', { name: 'Community' }).click();
  await expect(page.getByText('Q&A and threaded travel/provider discussion.')).toBeVisible();

  await page.getByLabel('Title').fill('E2E thread');
  await page.getByLabel('Body').fill('Can I check in after 6 PM?');
  await page.getByRole('button', { name: 'Publish' }).click();
  await expect(page.getByText('E2E thread')).toBeVisible();

  // Traveler should not see operations/admin-only analytics route/menu.
  await expect(page.getByRole('button', { name: 'Analytics' })).toHaveCount(0);
});
