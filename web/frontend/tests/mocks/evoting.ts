import Worker0 from './../json/api/proxies/dela-worker-0.json'
import Worker1 from './../json/api/proxies/dela-worker-1.json'
import Worker2 from './../json/api/proxies/dela-worker-2.json'
import Worker3 from './../json/api/proxies/dela-worker-3.json'

export const FORMID = 'b63bcb854121051f2d8cff04bf0ac9b524b534b704509a16a423448bde3321b4';

export async function mockFormsFormID(page: page, formStatus: number) {
  // clear current mock
  await page.unroute(`${process.env.DELA_PROXY_URL}/evoting/forms/${FORMID}`);
  await page.route(`${process.env.DELA_PROXY_URL}/evoting/forms/${FORMID}`, async (route) => {
    let formFile = '';
    switch (formStatus) {
      case 0:
        formFile = 'created.json';
        break;
      case 1:
        formFile = 'open.json';
        break;
      case 2:
        formFile = 'closed.json';
        break;
      case 3:
        formFile = 'shuffled.json';
        break;
      case 4:
        formFile = 'decrypted.json';
        break;
      case 5:
        formFile = 'combined.json';
        break;
    }
    await route.fulfill({
      path: `./tests/json/evoting/forms/${formFile}`,
    });
  });
}

export async function mockDKGActors(page: page, formStatus: number, initialized?: boolean) {
  // the nodes must have been initialized if the form changed state
  initialized = (initialized || formStatus > 0);
  for (const worker of [Worker0, Worker1, Worker2, Worker3]) {
    await page.route(`${worker.Proxy}/evoting/services/dkg/actors/${FORMID}`, async (route) => {
      let dkgActorsFile = '';
      switch (formStatus) {
        case 0:
          dkgActorsFile = initialized ? 'initialized.json' : 'uninitialized.json';
          break;
        case 1:
          dkgActorsFile = 'setup.json';
          break;
        case 6:
          dkgActorsFile = 'certified.json';
          break;
      }
      await route.fulfill({
        status: initialized ? 200 : 400,
        path: `./tests/json/evoting/dkgActors/${dkgActorsFile}`,
      });
    });
  }
}
