import { expect, test } from '@playwright/test';
import { default as i18n } from 'i18next';
import { assertHasFooter, assertHasNavBar, initI18n, logIn, setUp, translate } from './shared';
import { SCIPER_ADMIN, SCIPER_USER, mockPersonalInfo } from './mocks/api';
import { FORMID, mockFormsFormID, mockDKGActors } from './mocks/evoting';
import { mockProxies } from './mocks/api';

initI18n();

// main elements

test.beforeEach(async ({ page }) => {
  // mock empty list per default
  await mockFormsFormID(page, 0);
  for (const i of [0, 1, 2, 3]) {
    await mockProxies(page, i);
  }
  await mockDKGActors(page, 0, true);
  await mockPersonalInfo(page);
  await setUp(page, `/forms/${FORMID}`);
});

test('Assert navigation bar is present', async ({ page }) => {
  await assertHasNavBar(page);
});

test('Assert footer is present', async ({ page }) => {
  await assertHasFooter(page);
});
