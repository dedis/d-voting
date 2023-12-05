import { default as i18n } from 'i18next';
import { test, expect } from '@playwright/test';
import {
  initI18n,
  logIn,
  logOut,
  assertOnlyVisibleToAuthenticated,
  assertOnlyVisibleToAdmin,
} from './shared';
import { mockPersonalInfo } from './mocks';

initI18n();

test.beforeEach(async ({ page }) => {
  await mockPersonalInfo(page);
  await page.goto(`${process.env.FRONT_END_URL}/about`);
});

// unauthenticated

test('Assert D-Voting logo is present', async({ page }) => {
  const logo = await page.getByAltText(i18n.t('Workflow'));
  await expect(logo).toBeVisible();
  await logo.click();
  await expect(page).toHaveURL(process.env.FRONT_END_URL);
});

test('Assert link to form table is present', async({ page }) => {
  const forms = await page.getByRole('link', { name: i18n.t('navBarStatus') });
  await expect(forms).toBeVisible();
  await forms.click();
  await expect(page).toHaveURL(`${process.env.FRONT_END_URL}/form/index`);
});

// authenticated non-admin

test('Assert "Profile" button is visible upon logging in', async({ page }) => {
  await assertOnlyVisibleToAuthenticated(
    page, page.getByRole('button', { name: i18n.t('Profile') })
  );
});

// admin

test('Assert "Create form" button is (only) visible to admin', async({ page }) => {
  await assertOnlyVisibleToAdmin(
    page, page.getByRole('link', { name: i18n.t('navBarCreateForm')})
  );
});

test('Assert "Admin" button is (only) visible to admin', async({ page }) => {
  await assertOnlyVisibleToAdmin(
    page, page.getByRole('link', { name: i18n.t('navBarAdmin') })
  );
});
