import { FORMID } from './shared';
export const SCIPER_ADMIN = '123456';
export const SCIPER_OTHER_ADMIN = '987654';
export const SCIPER_USER = '789012';
export const SCIPER_OTHER_USER = '654321';

// /api/evoting

export async function mockDKGActors(page: page) {
  await page.route('/api/evoting/services/dkg/actors', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({ status: 200 });
    }
  });
}

export async function mockDKGActorsFormID(page: page) {
  await page.route(`/api/evoting/services/dkg/actors/${FORMID}`, async (route) => {
    if (route.request().method() === 'PUT') {
      await route.fulfill({ status: 200 });
    }
  });
}

export async function mockServicesShuffle(page: page) {
  await page.route('/api/evoting/services/shuffle/', async (route) => {
    if (route.request().method() === 'PUT') {
      await route.fulfill({ status: 200 });
    }
  });
}

export async function mockForms(page: page) {
  await page.route(`/api/evoting/forms/${FORMID}`, async (route) => {
    if (route.request().method() === 'PUT') {
      await route.fulfill({ status: 200 });
    }
  });
}

export async function mockFormsVote(page: page) {
  await page.route(`/api/evoting/forms/${FORMID}/vote`, async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: {
          Status: 0,
          Token:
            'eyJTdGF0dXMiOjAsIlRyYW5zYWN0aW9uSUQiOiJQQWluaEVjNVNzM2JiVWkxbldNWU55dWdPVkFpdVZ3YklZcGpKTFJ1SUdnPSIsIkxhc3RCbG9ja0lkeCI6NSwiVGltZSI6MTcwODAxNDgyMSwiSGFzaCI6ImtVT3g3Ykw0eC9IYXdwanppTityTVFIL3Fmb1pnRHBFUFc3S2tzRWl1TTA9IiwiU2lnbmF0dXJlIjoiZXlKT1lXMWxJam9pUWt4VExVTlZVbFpGTFVKT01qVTJJaXdpUkdGMFlTSTZJbU00UjFwUVYyWjZTRlpqT1dGV1dWaGhTbEUwWVhKT1UyRTRVbXB2VEVOSGIyZFBlbWRpVDFKSlYxVnVSbkE0ZFdaemQyMWlVVXA2ZWpScWJHNVFRa3QzUzJwWmVEaDJkVmgzZUhCWE5FeE1hVFZWUkRsUlBUMGlmUT09In0=',
        },
      });
    }
  });
}

// /api/config

export async function mockProxy(page: page) {
  await page.route('/api/config/proxy', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'text/html',
      body: `${process.env.DELA_PROXY_URL}`,
      headers: {
        'set-cookie':
          'connect.sid=s%3A5srES5h7hQ2fN5T71W59qh3cUSQL3Mix.fPoO3rOxui8yfTG7tFd7RPyasaU5VTkhxgdzVRWJyNk',
      },
    });
  });
}

// /api

export async function mockProxies(page: page, workerNumber: number) {
  await page.route(
    `/api/proxies/grpc%3A%2F%2Fdela-worker-${workerNumber}%3A2000`,
    async (route) => {
      if (route.request().method() === 'OPTIONS') {
        await route.fulfill({
          status: 200,
          headers: {
            'Access-Control-Allow-Headers': '*',
            'Access-Control-Allow-Origin': '*',
          },
        });
      } else {
        await route.fulfill({
          path: `./tests/json/api/proxies/dela-worker-${workerNumber}.json`,
        });
      }
    }
  );
}

export async function mockPersonalInfo(page: page, sciper?: string) {
  // clear current mock
  await page.unroute('/api/personal_info');
  await page.route('/api/personal_info', async (route) => {
    if (sciper) {
      route.fulfill({ path: `./tests/json/api/personal_info/${sciper}.json` });
    } else {
      route.fulfill({ status: 401, contentType: 'text/html', body: 'Unauthenticated' });
    }
  });
}

export async function mockGetDevLogin(page: page) {
  await page.route(`/api/get_dev_login/${SCIPER_ADMIN}`, async (route) => {
    await route.fulfill({});
  });
  await page.route(`/api/get_dev_login/${SCIPER_USER}`, async (route) => {
    await route.fulfill({});
  });
  if (
    process.env.REACT_APP_SCIPER_ADMIN !== undefined &&
    process.env.REACT_APP_SCIPER_ADMIN !== SCIPER_ADMIN
  ) {
    // dummy route for "Login" button depending on local configuration
    await page.route(`/api/get_dev_login/${process.env.REACT_APP_SCIPER_ADMIN}`, async (route) => {
      await route.fulfill({});
    });
  }
}

export async function mockLogout(page: page) {
  await page.route('/api/logout', async (route) => {
    await route.fulfill({});
  });
}

export async function mockAddRole(page: page) {
  await page.route('/api/add_role', async (route) => {
    await route.fulfill({ status: 200 });
  });
}
