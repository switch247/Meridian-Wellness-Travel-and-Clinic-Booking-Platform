import { test, expect } from '@playwright/test';

test.describe('Admin Flow', () => {
  test('Admin can assign roles and view audits', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Sign In');
    await page.getByLabel('Username').fill('admin');
    await page.getByLabel('Password').fill('Password123!');
    await page.click('button[type="submit"]');
    await page.getByRole('button', { name: 'Admin' }).click();
    await page.getByRole('button', { name: 'Assign Role' }).first().click();
    // MUI Select workaround: open combobox and click option
    // MUI Select workaround: open combobox and click option (force for visibility)
    await page.getByRole('combobox', { name: /Role/ }).click({ force: true });
    await page.getByRole('option', { name: /^traveler$/i }).click({ force: true });
    await page.getByRole('button', { name: /Assign(ing)?/ }).click();
    // Wait for dialog to close (combobox gone)
    await expect(page.getByRole('dialog', { name: 'Assign Role' })).not.toBeVisible();
    await page.getByRole('button', { name: 'Role Audits' }).click();
    await expect(page.getByRole('heading', { name: 'Role Audits' })).toBeVisible();
  });
});
