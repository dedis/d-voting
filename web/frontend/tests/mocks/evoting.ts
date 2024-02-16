import Worker0 from './../json/api/proxies/dela-worker-0.json';
import Worker1 from './../json/api/proxies/dela-worker-1.json';
import Worker2 from './../json/api/proxies/dela-worker-2.json';
import Worker3 from './../json/api/proxies/dela-worker-3.json';
import { FORMID } from './shared';

export async function mockForms(page: page, empty: boolean = true) {
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
        path: './tests/json/formIndex.json',
      });
    }
  });
}

export async function mockFormsFormID(page: page, formStatus: number) {
  // clear current mock
  await page.unroute(`${process.env.DELA_PROXY_URL}/evoting/forms/${FORMID}`);
  await page.route(`${process.env.DELA_PROXY_URL}/evoting/forms/${FORMID}`, async (route) => {
    const formFile = [
      'created.json',
      'open.json',
      'closed.json',
      'shuffled.json',
      'decrypted.json',
      '/result/clear.json', // default results
      'canceled.json',
      '/result/tie.json', // alternative results
    ][formStatus];
    await route.fulfill({
      path: `./tests/json/evoting/forms/${formFile}`,
    });
  });
}

export async function mockDKGActors(page: page, dkgActorsStatus: number, initialized: boolean) {
  for (const worker of [Worker0, Worker1, Worker2, Worker3]) {
    await page.route(`${worker.Proxy}/evoting/services/dkg/actors/${FORMID}`, async (route) => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          headers: {
            'Access-Control-Allow-Headers': '*',
            'Access-Control-Allow-Origin': '*',
          },
        });
      } else {
        let dkgActorsFile = '';
        switch (dkgActorsStatus) {
          case 0:
            dkgActorsFile = initialized ? 'initialized.json' : 'uninitialized.json';
            break;
          case 6:
            dkgActorsFile = worker === Worker0 ? 'setup.json' : 'certified.json'; // one node is set up, the remaining nodes are certified
            break;
        }
        await route.fulfill({
          status: initialized ? 200 : 400,
          path: `./tests/json/evoting/dkgActors/${dkgActorsFile}`,
        });
      }
    });
  }
}
