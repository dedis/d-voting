import { test } from '@playwright/test';
import { assertHasFooter, assertHasNavBar, initI18n, setUp } from './shared';
import { UPDATE, mockPersonalInfo } from './mocks';

initI18n();

test.beforeEach(async ({ page }) => {
  if (UPDATE === true) {
    return;
  }
  await mockPersonalInfo(page);
  await setUp(page, '/form/index');
});

test('Assert navigation bar is present', async ({ page }) => {
  test.skip(UPDATE === true, 'Do not run regular tests when updating HAR files');
  await assertHasNavBar(page);
});

test('Assert footer is present', async ({ page }) => {
  test.skip(UPDATE === true, 'Do not run regular tests when updating HAR files');
  await assertHasFooter(page);
});
