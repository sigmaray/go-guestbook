import { expect, test } from '@playwright/test';

import {
  acceptNextDialog,
  createAdminMessage,
  createEntity,
  createPublicMessage,
  executeSQL,
  login,
  truncateTable,
  uniqueId,
} from './helpers';

test.describe.serial('messages', () => {
  test.beforeEach(async ({ request }) => {
    await truncateTable(request, 'messages');
  });

  test('visitor can post and see a message', async ({ page }) => {
    const author = uniqueId('guest');
    const content = `Hello from ${author}`;

    await createPublicMessage(page, { author, content });

    await expect(page.locator('.message-author', { hasText: author })).toBeVisible();
    await expect(page.locator('.message-content', { hasText: content })).toBeVisible();
  });

  test('allows empty author on public form', async ({ page }) => {
    const content = `Anonymous message ${uniqueId('anon')}`;

    await createPublicMessage(page, { content });

    await expect(page.locator('.message-author', { hasText: 'Anonymous' })).toBeVisible();
    await expect(page.locator('.message-content', { hasText: content })).toBeVisible();
  });

  test('stores visitor metadata on public messages', async ({ page, request }) => {
    const content = `Metadata message ${uniqueId('meta')}`;

    await createPublicMessage(page, { content });

    const result = await executeSQL(
      request,
      `SELECT ip_address, user_agent, referer, accept_language, content FROM messages WHERE content = '${content}'`,
    );

    expect(result.rows).toHaveLength(1);
    expect(String(result.rows[0].ip_address || '')).not.toEqual('');
    expect(String(result.rows[0].user_agent || '')).not.toEqual('');
    expect(String(result.rows[0].accept_language || '')).not.toEqual('');
  });

  test('admin can view a message without editing', async ({ page }) => {
    const author = uniqueId('view-author');
    const content = `Viewable message ${author}`;

    await login(page);
    await createAdminMessage(page, { author, content, email: 'view@example.com' });

    await page.locator('.admin-messages__row').filter({ hasText: author }).getByRole('link', { name: 'View' }).click();

    await expect(page).toHaveURL(/\/admin\/messages\/\d+\/?$/);
    await expect(page.getByRole('heading', { name: /view message/i })).toBeVisible();
    await expect(page.locator('.message-author', { hasText: author })).toBeVisible();
    await expect(page.locator('.message-email', { hasText: 'view@example.com' })).toBeVisible();
    await expect(page.locator('.message-content', { hasText: content })).toBeVisible();
    await expect(page.locator('input, textarea')).toHaveCount(0);
  });

  test('admin can create, edit, and delete a message', async ({ page }) => {
    const author = uniqueId('admin-author');
    const content = `Admin created ${author}`;

    await login(page);
    await createAdminMessage(page, { author, content, email: 'admin@example.com' });

    await expect(page.locator('.message-author', { hasText: author })).toBeVisible();

    await page.getByRole('link', { name: /edit/i }).first().click();
    const updatedContent = `${content} updated`;
    await page.locator('#content').fill(updatedContent);
    await page.getByRole('button', { name: /update message/i }).click();

    await expect(page).toHaveURL(/\/admin\/messages\/?$/);
    await expect(page.locator('.message-content', { hasText: updatedContent })).toBeVisible();

    acceptNextDialog(page);
    await page.getByRole('button', { name: /delete/i }).first().click();

    await expect(page).toHaveURL(/\/admin\/messages\/?$/);
    await expect(page.locator('.message-author', { hasText: author })).toHaveCount(0);
  });

  test('admin messages list uses /admin/messages and paginates', async ({ page, request }) => {
    const prefix = uniqueId('page-msg');
    for (let i = 1; i <= 11; i += 1) {
      await createEntity(request, 'messages', {
        author: `${prefix}-${i}`,
        content: `Pagination message ${prefix}-${i}`,
      });
    }

    await login(page);
    await page.goto('/admin/messages', { waitUntil: 'domcontentloaded' });

    await expect(page).toHaveURL(/\/admin\/messages\/?$/);
    await expect(page.getByRole('link', { name: 'Messages', exact: true })).toHaveAttribute(
      'href',
      '/admin/messages',
    );
    await expect(page.getByRole('navigation', { name: 'Messages pagination' })).toBeVisible();
    await expect(page.locator('.admin-messages__row')).toHaveCount(10);

    await page.getByRole('link', { name: 'Next' }).click();
    await expect(page).toHaveURL(/\/admin\/messages\?page=2/);
    await expect(page.locator('.admin-messages__row')).toHaveCount(1);
    await expect(page.locator('.message-author', { hasText: `${prefix}-1` })).toBeVisible();
  });
});
