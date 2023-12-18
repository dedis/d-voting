import { default as i18n } from 'i18next';
import { expect, test } from '@playwright/test';
import {
  assertOnlyVisibleToAdmin,
  assertOnlyVisibleToAuthenticated,
  initI18n,
  logIn,
  setUp,
} from './shared';
import { SCIPER_USER, mockLogout, mockPersonalInfo } from './mocks';

initI18n();

test.beforeEach(async ({ page }) => {
  await mockPersonalInfo(page);
  await setUp(page, '/about');
});

// unauthenticated

test('Assert cookie is set', async ({ page }) => {
  const cookies = await page.context().cookies();
  await expect(cookies.find((cookie) => cookie.name === 'connect.sid')).toBeTruthy();
});

test('Assert D-Voting logo is present', async ({ page }) => {
  const logo = await page.getByAltText(i18n.t('Workflow'));
  await expect(logo).toBeVisible();
  await logo.click();
  await expect(page).toHaveURL('/');
});

test('Assert link to form table is present', async ({ page }) => {
  const forms = await page.getByRole('link', { name: i18n.t('navBarStatus') });
  await expect(forms).toBeVisible();
  await forms.click();
  await expect(page).toHaveURL('/form/index');
});

test('Assert "Login" button calls login API', async ({ page }) => {
  page.waitForRequest(new RegExp('/api/get_dev_login/[0-9]{6}'));
  await page.getByRole('button', { name: i18n.t('login') }).click();
});

// authenticated non-admin

test('Assert "Profile" button is visible upon logging in', async ({ page }) => {
  await assertOnlyVisibleToAuthenticated(
    page,
    page.getByRole('button', { name: i18n.t('Profile') })
  );
});

test('Assert "Logout" calls logout API', async ({ page, baseURL }) => {
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
  await assertOnlyVisibleToAdmin(
    page,
    page.getByRole('link', { name: i18n.t('navBarCreateForm') })
  );
});

test('Assert "Admin" button is (only) visible to admin', async ({ page }) => {
  await assertOnlyVisibleToAdmin(page, page.getByRole('link', { name: i18n.t('navBarAdmin') }));
});
