import { expect, test } from '@playwright/test';
import { assertHasFooter, assertHasNavBar, initI18n, logIn, setUp } from './shared';
import {
  SCIPER_ADMIN,
  SCIPER_OTHER_ADMIN,
  SCIPER_USER,
  mockPersonalInfo,
  mockProxies,
} from './mocks/api';
import { FORMID, mockDKGActors, mockFormsFormID } from './mocks/evoting';

initI18n();

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
  await mockPersonalInfo(page);
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

async function assertIsOnlyVisibleToOwnerStates(page: page, locator: locator, states: Array) {
  for (const i of states) {
    await test.step(`Assert is only visible to owner in state ${i}`, async () => {
      await setUpMocks(page, i, 6);
      await page.reload({ waitUntil: 'networkidle' });
      await assertIsOnlyVisibleToOwner(page, locator);
    });
  }
}

test('Assert "Add voters" button is only visible to owner', async ({ page }) => {
  await assertIsOnlyVisibleToOwnerStates(page, page.getByTestId('addVotersButton'), [0, 1]);
});