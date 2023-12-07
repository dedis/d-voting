import { default as i18n } from 'i18next';
import { test, expect } from '@playwright/test';
import {
  initI18n,
  setUp,
  logIn,
  logOut,
  assertOnlyVisibleToAuthenticated,
  assertOnlyVisibleToAdmin,
} from './shared';
import { SCIPER_ADMIN, SCIPER_USER, UPDATE, mockPersonalInfo } from './mocks';

initI18n();

test.beforeEach(async ({ page }) => {
  if (UPDATE === true) {
    return;
  }
  await setUp(page, `${process.env.FRONT_END_URL}`);
});

// helper tests to update related HAR files

test('Assert anonymous user HAR files are up-to-date', async({ page }) => {
  // comment the next line to update HAR files
  test.skip(UPDATE == false, 'Do not update HAR files');
  await mockPersonalInfo(page, '');
  await setUp(page, `${process.env.FRONT_END_URL}/about`);
});

test('Assert non-admin user HAR files are up-to-date', async({ page }) => {
  // comment the next line to update HAR files
  test.skip(UPDATE == false, 'Do not update HAR files');
  await mockPersonalInfo(page, SCIPER_USER);
  await page.context().request.get(`${process.env.FRONT_END_URL}/api/get_dev_login/${SCIPER_USER}`);
  await setUp(page, `${process.env.FRONT_END_URL}/about`);
});

test('Assert admin user HAR files are up-to-date', async({ page }) => {
  // comment the next line to update HAR files
  test.skip(UPDATE == false, 'Do not update HAR files');
  await mockPersonalInfo(page, SCIPER_ADMIN);
  await page.context().request.get(`${process.env.FRONT_END_URL}/api/get_dev_login/${SCIPER_ADMIN}`);
  await setUp(page, `${process.env.FRONT_END_URL}/about`);
});

// unauthenticated

test('Assert D-Voting logo is present', async({ page }) => {
  test.skip(UPDATE == true, 'Do not run regular tests when updating HAR files');
  const logo = await page.getByAltText(i18n.t('Workflow'));
  await expect(logo).toBeVisible();
  await logo.click();
  await expect(page).toHaveURL(process.env.FRONT_END_URL);
});

test('Assert link to form table is present', async({ page }) => {
  test.skip(UPDATE == true, 'Do not run regular tests when updating HAR files');
  const forms = await page.getByRole('link', { name: i18n.t('navBarStatus') });
  await expect(forms).toBeVisible();
  await forms.click();
  await expect(page).toHaveURL(`${process.env.FRONT_END_URL}/form/index`);
});

// authenticated non-admin

test('Assert "Profile" button is visible upon logging in', async({ page }) => {
  test.skip(UPDATE == true, 'Do not run regular tests when updating HAR files');
  await assertOnlyVisibleToAuthenticated(
    page, page.getByRole('button', { name: i18n.t('Profile') })
  );
});

// admin

test('Assert "Create form" button is (only) visible to admin', async({ page }) => {
  test.skip(UPDATE == true, 'Do not run regular tests when updating HAR files');
  await assertOnlyVisibleToAdmin(
    page, page.getByRole('link', { name: i18n.t('navBarCreateForm')})
  );
});

test('Assert "Admin" button is (only) visible to admin', async({ page }) => {
  test.skip(UPDATE == true, 'Do not run regular tests when updating HAR files');
  await assertOnlyVisibleToAdmin(
    page, page.getByRole('link', { name: i18n.t('navBarAdmin') })
  );
});
