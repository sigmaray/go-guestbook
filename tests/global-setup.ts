import { execFileSync } from 'node:child_process';
import path from 'node:path';

const guestbookPort =
  process.env.SKIP_DOCKER_SETUP === '1'
    ? (process.env.GO_GUESTBOOK_HTTP_PORT ?? '8084')
    : (process.env.GO_GUESTBOOK_HTTP_PORT ?? '18084');

const baseURL = `http://127.0.0.1:${guestbookPort}`;

async function waitForServer(timeoutMs = 120_000): Promise<boolean> {
  const deadline = Date.now() + timeoutMs;

  while (Date.now() < deadline) {
    try {
      const response = await fetch(baseURL, { redirect: 'follow' });
      if (response.ok) {
        return true;
      }
    } catch {
      // Server not ready yet.
    }

    await new Promise((resolve) => setTimeout(resolve, 1000));
  }

  return false;
}

export default async function globalSetup() {
  if (process.env.SKIP_DOCKER_SETUP === '1') {
    console.log('[go-guestbook setup] SKIP_DOCKER_SETUP=1, waiting for existing server');
    if (!(await waitForServer())) {
      throw new Error(`go-guestbook is not reachable at ${baseURL}`);
    }
    return;
  }

  if (await waitForServer(3_000)) {
    console.log(`[go-guestbook setup] Server already running at ${baseURL}`);
    return;
  }

  const projectRoot = path.resolve(__dirname, '..');
  const composeFile = path.join(projectRoot, 'docker-compose.test.yml');

  console.log('[go-guestbook setup] Starting test stack with Docker Compose...');
  execFileSync(
    'docker',
    ['compose', '-f', composeFile, 'up', '-d', '--build', '--wait'],
    {
      cwd: projectRoot,
      stdio: 'inherit',
      env: {
        ...process.env,
        GO_GUESTBOOK_HTTP_PORT: guestbookPort,
      },
    },
  );

  execFileSync(
    'docker',
    ['compose', '-f', composeFile, 'exec', '-T', 'app', './guestbook', 'migrate'],
    {
      cwd: projectRoot,
      stdio: 'inherit',
      env: {
        ...process.env,
        GO_GUESTBOOK_HTTP_PORT: guestbookPort,
      },
    },
  );

  if (!(await waitForServer())) {
    throw new Error(`go-guestbook test stack failed to become ready at ${baseURL}`);
  }

  execFileSync(
    'docker',
    ['compose', '-f', composeFile, 'exec', '-T', 'app', './guestbook', 'users-seed'],
    {
      cwd: projectRoot,
      stdio: ['pipe', 'inherit', 'inherit'],
      input: 'y\n',
      env: {
        ...process.env,
        GO_GUESTBOOK_HTTP_PORT: guestbookPort,
      },
    },
  );

  console.log(`[go-guestbook setup] Test stack ready at ${baseURL}`);
}
