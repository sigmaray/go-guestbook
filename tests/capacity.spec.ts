import { expect, test } from '@playwright/test';

import { createEntity, truncateTable, uniqueId } from './helpers';

const testMaxMessages = 5;

test.describe('message capacity', () => {
  test.beforeEach(async ({ request }) => {
    await truncateTable(request, 'messages');
  });

  test('rejects public posts when the guestbook is full', async ({ page, request }) => {
    for (let i = 1; i <= testMaxMessages; i += 1) {
      await createEntity(request, 'messages', {
        author: `fill-${i}`,
        content: `capacity filler ${i}`,
      });
    }

    const content = `Overflow message ${uniqueId('full')}`;
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await page.locator('#content').fill(content);
    await page.getByRole('button', { name: /post message/i }).click();

    await expect(page.locator('.guestbook-form__error')).toContainText(
      'maximum number of messages',
    );
    await expect(page.locator('.message-content', { hasText: content })).toHaveCount(0);
  });
});
