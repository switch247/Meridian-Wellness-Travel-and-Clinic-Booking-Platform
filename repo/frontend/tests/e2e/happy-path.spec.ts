import { expect, test } from '@playwright/test';

test('traveler happy path: login, catalog, community, analytics schedule guarded', async ({ page }) => {
  await page.goto('/login');

  const threadTitle = `E2E thread ${Date.now()}`;

  const username = `e2e_${Date.now()}`;
  await page.getByLabel('Username').fill(username);
  await page.getByLabel('Password').fill('Strong#Pass123');
  await page.getByRole('button', { name: 'Quick Register + Sign In' }).click();

  await expect(page.getByRole('heading', { name: 'Operational Dashboard' })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Traveler Snapshot' })).toBeVisible();

  await page.getByRole('button', { name: 'My Reservations' }).click();
  await expect(page.getByRole('heading', { name: 'My Reservations' })).toBeVisible();

  await page.getByRole('button', { name: 'Catalog' }).click();
  await expect(page.getByRole('heading', { name: 'Wellness Catalog' })).toBeVisible();

  await page.getByRole('button', { name: 'Community' }).click();
  await expect(page.getByRole('heading', { name: 'Community Discussions' })).toBeVisible();

  await page.getByRole('button', { name: 'New Discussion' }).click();
  await page.getByLabel('Title').fill(threadTitle);
  await page.getByLabel('Description').fill('Can I check in after 6 PM?');
  await page.getByRole('button', { name: 'Post Discussion' }).click();
  await expect(page.getByText(threadTitle)).toBeVisible();

  // Traveler should not see operations/admin-only analytics route/menu.
  await expect(page.getByRole('button', { name: 'Analytics' })).toHaveCount(0);
});
