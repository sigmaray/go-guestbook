import { expect, test } from '@playwright/test';

test.describe('test api', () => {
  test('creates entities and executes sql', async ({ request }) => {
    await request.post('/api/test/truncate', { data: { table: 'messages' } });

    const createResponse = await request.post('/api/test/entities', {
      data: {
        table: 'messages',
        values: {
          author: 'API Guest',
          email: 'api@example.com',
          content: 'Created through test API',
        },
      },
    });
    expect(createResponse.status()).toBe(201);

    const sqlResponse = await request.post('/api/test/sql', {
      data: {
        query: "SELECT author, content FROM messages WHERE author = 'API Guest'",
      },
    });
    expect(sqlResponse.ok()).toBeTruthy();

    const body = await sqlResponse.json();
    expect(body.rows).toEqual([
      {
        author: 'API Guest',
        content: 'Created through test API',
      },
    ]);
  });
});
