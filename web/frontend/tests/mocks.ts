export const SCIPER_ADMIN = process.env.REACT_APP_SCIPER_ADMIN;
export const SCIPER_USER = '789012';

export async function mockGetDevLogin (page: any) {
  await page.routeFromHAR(
    `./tests/hars/get_dev_login.admin.har`,
    {
      url: `${process.env.FRONT_END_URL}/api/get_dev_login/${SCIPER_ADMIN}`,
      update: false,
    });
  await page.routeFromHAR(
    `./tests/hars/get_dev_login.${SCIPER_USER}.har`,
    {
      url: `${process.env.FRONT_END_URL}/api/get_dev_login/${SCIPER_USER}`,
      update: false,
    });
}

export async function mockLogout (page: any) {
  await page.routeFromHAR(
    './tests/hars/logout.har',
    {
      url: `${process.env.FRONT_END_URL}/api/logout`,
      update: false,
    });
}
