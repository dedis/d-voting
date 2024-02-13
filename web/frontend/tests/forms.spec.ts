import { expect, test } from '@playwright/test';
import { default as i18n } from 'i18next';
import { assertHasFooter, assertHasNavBar, initI18n, logIn, setUp } from './shared';
import {
  SCIPER_ADMIN,
  SCIPER_OTHER_ADMIN,
  SCIPER_OTHER_USER,
  SCIPER_USER,
  mockDKGActors as mockAPIDKGActors,
  mockAddRole,
  mockDKGActorsFormID,
  mockForms,
  mockPersonalInfo,
  mockProxies,
  mockServicesShuffle,
} from './mocks/api';
import { mockDKGActors, mockFormsFormID } from './mocks/evoting';
import { FORMID } from './mocks/shared';
import Worker0 from './json/api/proxies/dela-worker-0.json';
import Worker1 from './json/api/proxies/dela-worker-1.json';
import Worker2 from './json/api/proxies/dela-worker-2.json';
import Worker3 from './json/api/proxies/dela-worker-3.json';

initI18n();

const prettyFormStates = [
  'Initial',
  'Open',
  'Closed',
  'ShuffledBallots',
  'PubSharesSubmitted',
  'ResultAvailable',
  'Canceled',
];

// main elements

async function setUpMocks(
  page: page,
  formStatus: number,
  dkgActorsStatus: number,
  initialized?: boolean
) {
  // the nodes must have been initialized if they changed state
  initialized = initialized || dkgActorsStatus > 0;
  await mockFormsFormID(page, formStatus);
  for (const i of [0, 1, 2, 3]) {
    await mockProxies(page, i);
  }
  await mockDKGActors(page, dkgActorsStatus, initialized);
  await mockAPIDKGActors(page);
  await mockPersonalInfo(page);
  await mockDKGActorsFormID(page);
  await mockServicesShuffle(page);
  await mockForms(page);
}

test.beforeEach(async ({ page }) => {
  // mock empty list per default
  setUpMocks(page, 0, 0, false);
  await setUp(page, `/forms/${FORMID}`);
});

test('Assert navigation bar is present', async ({ page }) => {
  await assertHasNavBar(page);
});

test('Assert footer is present', async ({ page }) => {
  await assertHasFooter(page);
});

async function assertIsOnlyVisibleToOwner(page: page, locator: locator) {
  await test.step('Assert is hidden to unauthenticated user', async () => {
    await expect(locator).toBeHidden();
  });
  await test.step('Assert is hidden to authenticated non-admin user', async () => {
    await logIn(page, SCIPER_USER);
    await expect(locator).toBeHidden();
  });
  await test.step('Assert is hidden to non-owner admin', async () => {
    await logIn(page, SCIPER_ADMIN);
    await expect(locator).toBeHidden();
  });
  await test.step('Assert is visible to owner admin', async () => {
    await logIn(page, SCIPER_OTHER_ADMIN);
    await expect(locator).toBeVisible();
  });
}

async function assertIsOnlyVisibleInStates(
  page: page,
  locator: locator,
  states: Array,
  assert: Function,
  dkgActorsStatus?: number,
  initialized?: boolean
) {
  for (const i of states) {
    await test.step(`Assert is visible in form state '${prettyFormStates.at(i)}'`, async () => {
      await setUpMocks(page, i, dkgActorsStatus === undefined ? 6 : dkgActorsStatus, initialized);
      await page.reload({ waitUntil: 'networkidle' });
      await assert(page, locator);
    });
  }
  for (const i of [0, 1, 2, 3, 4, 5].filter((x) => !states.includes(x))) {
    await test.step(
      `Assert is not visible in form state '${prettyFormStates.at(i)}'`,
      async () => {
        await setUpMocks(page, i, dkgActorsStatus === undefined ? 6 : dkgActorsStatus, initialized);
        await page.reload({ waitUntil: 'networkidle' });
        await expect(locator).toBeHidden();
      }
    );
  }
}

async function assertRouteIsCalled(
  page: page,
  url: string,
  key: string,
  action: string,
  formStatus: number,
  confirmation: boolean,
  dkgActorsStatus?: number,
  initialized?: boolean
) {
  await setUpMocks(
    page,
    formStatus,
    dkgActorsStatus === undefined ? 6 : dkgActorsStatus,
    initialized
  );
  await logIn(page, SCIPER_OTHER_ADMIN);
  page.waitForRequest(async (request) => {
    const body = await request.postDataJSON();
    return request.url() === url && request.method() === 'PUT' && body.Action === action;
  });
  await page.getByRole('button', { name: i18n.t(key) }).click();
  if (confirmation) {
    await page.getByRole('button', { name: i18n.t('yes') }).click();
  }
}

test('Assert "Add voters" button is only visible to owner', async ({ page }) => {
  await assertIsOnlyVisibleInStates(
    page,
    page.getByTestId('addVotersButton'),
    [0, 1],
    assertIsOnlyVisibleToOwner
  );
});

test('Assert "Add voters" button allows to add voters', async ({ page, baseURL }) => {
  await setUpMocks(page, 0, 6);
  await mockAddRole(page);
  await logIn(page, SCIPER_OTHER_ADMIN);
  // we expect one call per new voter
  for (const sciper of [SCIPER_OTHER_ADMIN, SCIPER_ADMIN, SCIPER_USER]) {
    page.waitForRequest(async (request) => {
      const body = await request.postDataJSON();
      return (
        request.url() === `${baseURL}/api/add_role` &&
        request.method() === 'POST' &&
        body.permission === 'vote' &&
        body.subject === FORMID &&
        body.userId.toString() === sciper
      );
    });
  }
  await page.getByTestId('addVotersButton').click();
  // menu should be visible
  const textbox = await page.getByRole('textbox', { name: 'SCIPERs' });
  await expect(textbox).toBeVisible();
  // add 3 voters (owner admin, non-owner admin, user)
  await textbox.fill(`${SCIPER_OTHER_ADMIN}\n${SCIPER_ADMIN}\n${SCIPER_USER}`);
  // click on confirmation
  await page.getByTestId('addVotersConfirm').click();
});

test('Assert "Initialize" button is only visible to owner', async ({ page }) => {
  await assertIsOnlyVisibleInStates(
    page,
    page.getByRole('button', { name: i18n.t('initialize') }),
    [0],
    assertIsOnlyVisibleToOwner,
    0,
    false
  );
});

test('Assert "Initialize" button calls route to initialize nodes', async ({ page, baseURL }) => {
  await setUpMocks(page, 0, 0, false);
  await logIn(page, SCIPER_OTHER_ADMIN);
  // we expect one call per worker node
  for (const worker of [Worker0, Worker1, Worker2, Worker3]) {
    page.waitForRequest(async (request) => {
      const body = await request.postDataJSON();
      return (
        request.url() === `${baseURL}/api/evoting/services/dkg/actors` &&
        request.method() === 'POST' &&
        body.FormID === FORMID &&
        body.Proxy === worker.Proxy
      );
    });
  }
  await page.getByRole('button', { name: i18n.t('initialize') }).click();
});

test('Assert "Setup" button is only visible to owner', async ({ page }) => {
  await assertIsOnlyVisibleInStates(
    page,
    page.getByRole('button', { name: i18n.t('setup') }),
    [0],
    assertIsOnlyVisibleToOwner,
    0,
    true
  );
});

test('Assert "Setup" button calls route to setup node', async ({ page, baseURL }) => {
  await setUpMocks(page, 0, 0, true);
  await logIn(page, SCIPER_OTHER_ADMIN);
  // we expect one call with the chosen worker node
  page.waitForRequest(async (request) => {
    const body = await request.postDataJSON();
    return (
      request.url() === `${baseURL}/api/evoting/services/dkg/actors/${FORMID}` &&
      request.method() === 'PUT' &&
      body.Action === 'setup' &&
      body.Proxy === Worker1.Proxy
    );
  });
  // open node selection window
  await page.getByRole('button', { name: i18n.t('setup') }).click();
  await expect(page.getByTestId('nodeSetup')).toBeVisible();
  // choose second worker node
  await page.getByLabel(Worker1.NodeAddr).check();
  // confirm
  await page.getByRole('button', { name: i18n.t('setupNode') }).click();
});

test('Assert "Open" button is only visible to owner', async ({ page }) => {
  await assertIsOnlyVisibleInStates(
    page,
    page.getByRole('button', { name: i18n.t('open') }),
    [0],
    assertIsOnlyVisibleToOwner
  );
});

test('Assert "Open" button calls route to open form', async ({ page, baseURL }) => {
  await assertRouteIsCalled(
    page,
    `${baseURL}/api/evoting/forms/${FORMID}`,
    'open',
    'open',
    0,
    false,
    6
  );
});

test('Assert "Cancel" button is only visible to owner', async ({ page }) => {
  await assertIsOnlyVisibleInStates(
    page,
    page.getByRole('button', { name: i18n.t('cancel') }),
    [1],
    assertIsOnlyVisibleToOwner
  );
});

test('Assert "Cancel" button calls route to cancel form', async ({ page, baseURL }) => {
  await assertRouteIsCalled(
    page,
    `${baseURL}/api/evoting/forms/${FORMID}`,
    'cancel',
    'cancel',
    1,
    true
  );
});

test('Assert "Close" button is only visible to owner', async ({ page }) => {
  await assertIsOnlyVisibleInStates(
    page,
    page.getByRole('button', { name: i18n.t('close') }),
    [1],
    assertIsOnlyVisibleToOwner
  );
});

test('Assert "Close" button calls route to close form', async ({ page, baseURL }) => {
  await assertRouteIsCalled(
    page,
    `${baseURL}/api/evoting/forms/${FORMID}`,
    'close',
    'close',
    1,
    true
  );
});

test('Assert "Shuffle" button is only visible to owner', async ({ page }) => {
  await assertIsOnlyVisibleInStates(
    page,
    page.getByRole('button', { name: i18n.t('shuffle') }),
    [2],
    assertIsOnlyVisibleToOwner
  );
});

test('Assert "Shuffle" button calls route to shuffle form', async ({ page, baseURL }) => {
  await assertRouteIsCalled(
    page,
    `${baseURL}/api/evoting/services/shuffle/${FORMID}`,
    'shuffle',
    'shuffle',
    2,
    false
  );
});

test('Assert "Decrypt" button is only visible to owner', async ({ page }) => {
  await assertIsOnlyVisibleInStates(
    page,
    page.getByRole('button', { name: i18n.t('decrypt') }),
    [3],
    assertIsOnlyVisibleToOwner
  );
});

test('Assert "Decrypt" button calls route to decrypt form', async ({ page, baseURL }) => {
  await assertRouteIsCalled(
    page,
    `${baseURL}/api/evoting/services/dkg/actors/${FORMID}`,
    'decrypt',
    'computePubshares',
    3,
    false
  );
});

test('Assert "Combine" button is only visible to owner', async ({ page }) => {
  await assertIsOnlyVisibleInStates(
    page,
    page.getByRole('button', { name: i18n.t('combine') }),
    [4],
    assertIsOnlyVisibleToOwner
  );
});

test('Assert "Combine" button calls route to combine form', async ({ page, baseURL }) => {
  await assertRouteIsCalled(
    page,
    `${baseURL}/api/evoting/forms/${FORMID}`,
    'combine',
    'combineShares',
    4,
    false
  );
});

test('Assert "Delete" button is only visible to owner', async ({ page }) => {
  test.setTimeout(60000); // Firefox is exceeding the default timeout on this test
  await assertIsOnlyVisibleInStates(
    page,
    page.getByRole('button', { name: i18n.t('delete') }),
    [0, 1, 2, 3, 4, 6],
    assertIsOnlyVisibleToOwner
  );
});

test('Assert "Delete" button calls route to delete form', async ({ page, baseURL }) => {
  for (const i of [0, 1, 2, 3, 4, 6]) {
    await setUpMocks(page, i, 6);
    await logIn(page, SCIPER_OTHER_ADMIN);
    page.waitForRequest(async (request) => {
      return (
        request.url() === `${baseURL}/api/evoting/forms/${FORMID}` && request.method() === 'DELETE'
      );
    });
    await page.getByRole('button', { name: i18n.t('delete') }).click();
    await page.getByRole('button', { name: i18n.t('yes') }).click();
  }
});

test('Assert "Vote" button is visible to admin/non-admin voter user', async ({ page }) => {
  await assertIsOnlyVisibleInStates(
    page,
    page.getByRole('button', { name: i18n.t('vote'), exact: true }), // by default name is not matched exactly which returns both the "Vote" and the "Add voters" button
    [1],
    // eslint-disable-next-line @typescript-eslint/no-shadow
    async function (page: page, locator: locator) {
      await test.step('Assert is hidden to unauthenticated user', async () => {
        await expect(locator).toBeHidden();
      });
      await test.step('Assert is hidden to authenticated non-voter user', async () => {
        await logIn(page, SCIPER_OTHER_USER);
        await expect(locator).toBeHidden();
      });
      await test.step('Assert is visible to authenticated voter user', async () => {
        await logIn(page, SCIPER_USER);
        await expect(locator).toBeVisible();
      });
      await test.step('Assert is hidden to non-voter admin', async () => {
        await logIn(page, SCIPER_OTHER_ADMIN);
        await expect(locator).toBeHidden();
      });
      await test.step('Assert is visible to voter admin', async () => {
        await logIn(page, SCIPER_ADMIN);
        await expect(locator).toBeVisible();
      });
    }
  );
});

test('Assert "Vote" button gets voting form', async ({ page }) => {
  await setUpMocks(page, 1, 6);
  await logIn(page, SCIPER_USER);
  page.waitForRequest(`${process.env.DELA_PROXY_URL}/evoting/forms/${FORMID}`);
  await page.getByRole('button', { name: i18n.t('vote') }).click();
  // go back to form management page
  await setUp(page, `/forms/${FORMID}`);
  await logIn(page, SCIPER_ADMIN);
  page.waitForRequest(`${process.env.DELA_PROXY_URL}/evoting/forms/${FORMID}`);
  await page.getByRole('button', { name: i18n.t('vote') }).click();
});
