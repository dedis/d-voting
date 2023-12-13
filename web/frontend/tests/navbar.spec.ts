import { default as i18n } from 'i18next';
import { expect, test } from '@playwright/test';
import {
  assertOnlyVisibleToAdmin,
  assertOnlyVisibleToAuthenticated,
  initI18n,
  logIn,
  setUp,
} from './shared';
import { SCIPER_ADMIN, SCIPER_USER, UPDATE, mockLogout, mockPersonalInfo } from './mocks';

initI18n();

test.beforeEach(async ({ page }) => {
  if (UPDATE === true) {
    return;
  }
  await mockPersonalInfo(page);
  await setUp(page, '/about');
});

// helper tests to update related HAR files

test('Assert anonymous user HAR files are up-to-date', async ({ page }) => {
  test.skip(UPDATE === false, 'Do not update HAR files');
  await mockPersonalInfo(page);
  await setUp(page, '/about');
});

test('Assert non-admin user HAR files are up-to-date', async ({ page }) => {
  test.skip(UPDATE === false, 'Do not update HAR files');
  await mockPersonalInfo(page, SCIPER_USER);
  await page.context().request.get(`/api/get_dev_login/${SCIPER_USER}`);
  await setUp(page, '/about');
});

test('Assert admin user HAR files are up-to-date', async ({ page }) => {
  test.skip(UPDATE === false, 'Do not update HAR files');
  await mockPersonalInfo(page, SCIPER_ADMIN);
  await page.context().request.get(`/api/get_dev_login/${SCIPER_ADMIN}`);
  await setUp(page, '/about');
});

// unauthenticated

test('Assert cookie is set', async ({ page }) => {
  test.skip(UPDATE === true, 'Do not run regular tests when updating HAR files');
  const cookies = await page.context().cookies();
  await expect(cookies.find((cookie) => cookie.name === 'connect.sid')).toBeTruthy();
});

test('Assert D-Voting logo is present', async ({ page }) => {
  test.skip(UPDATE === true, 'Do not run regular tests when updating HAR files');
  const logo = await page.getByAltText(i18n.t('Workflow'));
  await expect(logo).toBeVisible();
  await logo.click();
  await expect(page).toHaveURL('/');
});

test('Assert link to form table is present', async ({ page }) => {
  test.skip(UPDATE === true, 'Do not run regular tests when updating HAR files');
  const forms = await page.getByRole('link', { name: i18n.t('navBarStatus') });
  await expect(forms).toBeVisible();
  await forms.click();
  await expect(page).toHaveURL('/form/index');
});

test('Assert "Login" button calls login API', async ({ page }) => {
  test.skip(UPDATE === true, 'Do not run regular tests when updating HAR files');
  page.waitForRequest(new RegExp('/api/get_dev_login/[0-9]{6}'));
  await page.getByRole('button', { name: i18n.t('login') }).click();
});

// authenticated non-admin

test('Assert "Profile" button is visible upon logging in', async ({ page }) => {
  test.skip(UPDATE === true, 'Do not run regular tests when updating HAR files');
  await assertOnlyVisibleToAuthenticated(
    page,
    page.getByRole('button', { name: i18n.t('Profile') })
  );
});

test('Assert "Logout" calls logout API', async ({ page, baseURL }) => {
  test.skip(UPDATE === true, 'Do not run regular tests when updating HAR files');
  await mockLogout(page);
  await logIn(page, SCIPER_USER);
  page.waitForRequest(
    (request) => request.url() === `${baseURL}/api/logout` && request.method() === 'POST'
  );
  for (const [role, key] of [
    ['button', 'Profile'],
    ['menuitem', 'logout'],
    ['button', 'continue'],
  ]) {
    await page.getByRole(role, { name: i18n.t(key) }).click();
  }
});

// admin

test('Assert "Create form" button is (only) visible to admin', async ({ page }) => {
  test.skip(UPDATE === true, 'Do not run regular tests when updating HAR files');
  await assertOnlyVisibleToAdmin(
    page,
    page.getByRole('link', { name: i18n.t('navBarCreateForm') })
  );
});

test('Assert "Admin" button is (only) visible to admin', async ({ page }) => {
  test.skip(UPDATE === true, 'Do not run regular tests when updating HAR files');
  await assertOnlyVisibleToAdmin(page, page.getByRole('link', { name: i18n.t('navBarAdmin') }));
});
