export const SCIPER_ADMIN = '123456';
export const SCIPER_USER = '789012';

export async function mockPersonalInfo(page: page, sciper?: string) {
  // clear current mock
  await page.unroute('/api/personal_info');
  await page.route('/api/personal_info', async (route) => {
    if (sciper) {
      route.fulfill({ path: `./tests/json/personal_info/${sciper}.json` });
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

export async function mockEvoting(page: page, empty: boolean = true) {
  // clear current mock
  await page.unroute(`${process.env.DELA_PROXY_URL}/evoting/forms`);
  await page.route(`${process.env.DELA_PROXY_URL}/evoting/forms`, async (route) => {
    if (route.request().method() === 'OPTIONS') {
      await route.fulfill({
        status: 200,
        headers: {
          'Access-Control-Allow-Headers': '*',
          'Access-Control-Allow-Origin': '*',
        },
      });
    } else if (empty) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: '{"Forms": []}',
      });
    } else {
      await route.fulfill({
        path: './tests/json/formList.json',
      });
    }
  });
}
