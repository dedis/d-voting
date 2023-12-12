export const SCIPER_ADMIN = '123456';
export const SCIPER_USER = '789012';
export const UPDATE = false;

export async function mockPersonalInfo (page: any, sciper?: string) {
  // clear current mock
  await page.unroute('/api/personal_info');
  await page.routeFromHAR(
    `./tests/hars/${sciper ?? 'anonymous'}/personal_info.har`,
    {
      url: '/api/personal_info',
      update: UPDATE,
    });
}

export async function mockGetDevLogin (page: any) {
  await page.routeFromHAR(
    `./tests/hars/${SCIPER_ADMIN}/get_dev_login.har`,
    {
      url: `/api/get_dev_login/${SCIPER_ADMIN}`,
      update: UPDATE,
    });
  await page.routeFromHAR(
    `./tests/hars/${SCIPER_USER}/get_dev_login.har`,
    {
      url: `/api/get_dev_login/${SCIPER_USER}`,
      update: UPDATE,
    });
  if (process.env.REACT_APP_SCIPER_ADMIN !== undefined && process.env.REACT_APP_SCIPER_ADMIN !== SCIPER_ADMIN) {
    // dummy route for "Login" button depending on local configuration
    await page.route(
      `/api/get_dev_login/${process.env.REACT_APP_SCIPER_ADMIN}`,
      async route => {await route.fulfill({});}
    );
  }
}

export async function mockLogout (page: any) {
  await page.route(
    '/api/logout',
    async route => {await route.fulfill({});}
  );
}

export async function mockProxy (page: any) {
  await page.route(
    '/api/config/proxy',
    async route => {
      await route.fulfill(
        {
          status: 200,
          contentType: 'text/html',
          body: `${process.env.DELA_PROXY_URL}`,
          headers: {'set-cookie': 'connect.sid=s%3A5srES5h7hQ2fN5T71W59qh3cUSQL3Mix.fPoO3rOxui8yfTG7tFd7RPyasaU5VTkhxgdzVRWJyNk'},
        });
    }
  )
}
