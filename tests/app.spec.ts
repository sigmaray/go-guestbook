import { expect, test } from '@playwright/test';

import { login } from './helpers';

test.describe('public', () => {
  test('serves the homepage', async ({ page }) => {
    const response = await page.goto('/', { waitUntil: 'domcontentloaded' });

    expect(response?.ok()).toBeTruthy();
    await expect(page).toHaveTitle(/Go Guestbook/i);
    await expect(page.locator('body')).toContainText(/Go Guestbook/i);
    await expect(page.getByRole('button', { name: /post message/i })).toBeVisible();
  });

  test('shows Leave a Message below Messages', async ({ page }) => {
    await page.goto('/', { waitUntil: 'domcontentloaded' });

    const messagesHeading = page.getByRole('heading', { name: 'Messages', exact: true });
    const leaveHeading = page.getByRole('heading', { name: 'Leave a Message', exact: true });

    await expect(messagesHeading).toBeVisible();
    await expect(leaveHeading).toBeVisible();

    const messagesBox = await messagesHeading.boundingBox();
    const leaveBox = await leaveHeading.boundingBox();
    expect(messagesBox).not.toBeNull();
    expect(leaveBox).not.toBeNull();
    expect(messagesBox!.y).toBeLessThan(leaveBox!.y);
  });

  test('hides admin link from guests and shows it when logged in', async ({ page }) => {
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('link', { name: 'Admin' })).toHaveCount(0);

    await login(page);
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('link', { name: 'Admin' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Admin' })).toHaveAttribute('href', '/admin/messages');
  });

  test('serves health endpoint', async ({ request }) => {
    const response = await request.get('/health');
    expect(response.ok()).toBeTruthy();
    await expect(response.json()).resolves.toEqual({ status: 'ok' });
  });

  test('responds to HEAD / for uptime checks', async ({ request }) => {
    const response = await request.head('/');
    expect(response.ok()).toBeTruthy();
    expect(response.status()).toBe(200);
    expect(await response.body()).toEqual(Buffer.from(''));
  });

  test('redirects unauthenticated users from admin', async ({ page }) => {
    await page.goto('/admin/messages', { waitUntil: 'domcontentloaded' });
    await expect(page).toHaveURL(/\/login/);
  });

  test('redirects /admin/ to /admin/messages when logged in', async ({ page }) => {
    await login(page);
    await page.goto('/admin/', { waitUntil: 'domcontentloaded' });
    await expect(page).toHaveURL(/\/admin\/messages\/?$/);
  });
});

test.describe('auth', () => {
  test('login page is available', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });

    await expect(page.locator('#username')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.getByRole('button', { name: /login/i })).toBeVisible();
  });

  test('admin can log in and log out', async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toContainText(/Messages|Logout/i);

    await page.getByRole('button', { name: /logout/i }).click();
    await expect(page).toHaveURL(/\/(?:login)?$/);
  });

  test('rejects invalid credentials', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.locator('#username').fill('admin');
    await page.locator('#password').fill('wrong-password');
    await page.getByRole('button', { name: /login/i }).click();

    await expect(page.locator('.error')).toContainText('Invalid username or password');
    await expect(page).toHaveURL(/\/login/);
  });
});
