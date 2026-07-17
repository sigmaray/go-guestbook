import { expect, test } from '@playwright/test';
import { acceptNextDialog, login, truncateTable } from './helpers';

test.describe('Dev tools', () => {
  test.beforeEach(async ({ page, request }) => {
    await truncateTable(request, 'messages');
    await login(page);
  });

  test('shows tools page with clear, seed, and SQL panels', async ({ page }) => {
    await page.goto('/admin/tools', { waitUntil: 'domcontentloaded' });

    await expect(page.getByRole('heading', { name: 'Development Tools' })).toBeVisible();
    await expect(page.getByText('Available only when GO_GUESTBOOK_DEV_TOOLS_ENABLED is set.')).toBeVisible();
    await expect(page.getByRole('link', { name: 'Tools' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Clear table' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Seed messages' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Run query' })).toBeVisible();
  });

  test('seeds sample messages', async ({ page }) => {
    await page.goto('/admin/tools', { waitUntil: 'domcontentloaded' });

    await page.getByRole('button', { name: 'Seed messages' }).click();
    await expect(page.getByText(/Created \d+ message\(s\)\./)).toBeVisible();

    await page.goto('/admin/messages', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('cell', { name: 'Guest 1', exact: true })).toBeVisible();
  });

  test('runs a SELECT query and shows results', async ({ page }) => {
    await page.goto('/admin/tools', { waitUntil: 'domcontentloaded' });
    await page.getByRole('button', { name: 'Seed messages' }).click();
    await expect(page.getByText(/Created \d+ message\(s\)\./)).toBeVisible();

    await page.locator('#sql-query').fill('SELECT author FROM messages ORDER BY id LIMIT 1');
    await page.getByRole('button', { name: 'Run query' }).click();

    await expect(page.locator('.dev-tools__result-table')).toBeVisible();
    await expect(page.locator('.dev-tools__result-table')).toContainText('Guest 1');
  });

  test('clears the messages table after confirm', async ({ page }) => {
    await page.goto('/admin/tools', { waitUntil: 'domcontentloaded' });
    await page.getByRole('button', { name: 'Seed messages' }).click();
    await expect(page.getByText(/Created \d+ message\(s\)\./)).toBeVisible();

    await page.locator('#clear-table-select').selectOption('messages');
    acceptNextDialog(page);
    await page.getByRole('button', { name: 'Clear table' }).click();
    await expect(page.getByText('Table "messages" cleared.')).toBeVisible();

    await page.goto('/admin/messages', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('cell', { name: 'Guest 1', exact: true })).toHaveCount(0);
  });
});
