export async function mockPersonalInfo (page: any, sciper) {
  await page.routeFromHAR(
    sciper ? `./tests/hars/personal_info.${sciper}.har` : './tests/hars/personal_info.har',
    {
      url: `${process.env.FRONT_END_URL}/api/personal_info`,
      update: false,
    });
}

export async function mockGetDevLogin (page: any, sciper) {
  await page.routeFromHAR(
    `./tests/hars/get_dev_login.${sciper}.har`,
    {
      url: `${process.env.FRONT_END_URL}/api/get_dev_login/${sciper}`,
      update: false,
    });
  // dummy route for "Login" button depending on local configuration
  await page.route(
    `${process.env.FRONT_END_URL}/api/get_dev_login/${process.env.REACT_APP_SCIPER_ADMIN}`,
    async route => {await route.fulfill({});}
  );
}

export async function mockLogout (page: any, sciper) {
  await page.routeFromHAR(
    `./tests/hars/logout.${sciper}.har`,
    {
      url: `${process.env.FRONT_END_URL}/api/logout`,
      update: false,
    });
}
