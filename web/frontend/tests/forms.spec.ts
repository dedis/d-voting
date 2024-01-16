import { expect, test } from '@playwright/test';
import { assertHasFooter, assertHasNavBar, initI18n, setUp } from './shared';
import { mockPersonalInfo, mockProxies } from './mocks/api';
import { FORMID, mockDKGActors, mockFormsFormID } from './mocks/evoting';

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
