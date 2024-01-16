export async function mockProxies(page: page, workerNumber: number) {
  await page.route(`/api/proxies/grpc%3A%2F%2Fdela-worker-${workerNumber}%3A2000`, async (route) => {
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
  });
}

