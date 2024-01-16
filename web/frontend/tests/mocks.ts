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
