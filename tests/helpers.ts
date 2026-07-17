import { expect, type APIRequestContext, type Page } from '@playwright/test';

export function uniqueId(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export async function truncateTable(request: APIRequestContext, table: string) {
  const response = await request.post('/api/test/truncate', {
    data: { table },
  });
  expect(response.ok()).toBeTruthy();
}

export async function login(page: Page, username = 'admin', password = 'admin') {
  await page.goto('/login', { waitUntil: 'domcontentloaded' });
  await page.locator('#username').fill(username);
  await page.locator('#password').fill(password);
  await page.getByRole('button', { name: /login/i }).click();
  await expect(page).toHaveURL(/\/admin\/messages\/?$/);
}

export async function createPublicMessage(
  page: Page,
  opts: { author?: string; content: string; email?: string },
) {
  await page.goto('/', { waitUntil: 'domcontentloaded' });
  if (opts.author) {
    await page.locator('#author').fill(opts.author);
  }
  if (opts.email) {
    await page.locator('#email').fill(opts.email);
  }
  await page.locator('#content').fill(opts.content);
  await page.getByRole('button', { name: /post message/i }).click();
  await expect(page).toHaveURL('/');
}

export async function executeSQL(request: APIRequestContext, query: string) {
  const response = await request.post('/api/test/sql', {
    data: { query },
  });
  expect(response.ok()).toBeTruthy();
  return response.json() as Promise<{ rows: Array<Record<string, unknown>> }>;
}

export async function createEntity(
  request: APIRequestContext,
  table: string,
  values: Record<string, unknown>,
) {
  const response = await request.post('/api/test/entities', {
    data: { table, values },
  });
  expect(response.ok()).toBeTruthy();
  return response.json();
}

export async function createAdminMessage(
  page: Page,
  opts: { author: string; content: string; email?: string },
) {
  await page.goto('/admin/messages/new', { waitUntil: 'domcontentloaded' });
  await page.locator('#author').fill(opts.author);
  if (opts.email) {
    await page.locator('#email').fill(opts.email);
  }
  await page.locator('#content').fill(opts.content);
  await page.getByRole('button', { name: /create message/i }).click();
  await expect(page).toHaveURL(/\/admin\/messages\/?$/);
}

export async function acceptNextDialog(page: Page) {
  page.once('dialog', (dialog) => dialog.accept());
}
