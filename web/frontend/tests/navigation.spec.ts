import { default as i18n } from 'i18next';
import { test, expect } from '@playwright/test';
import {
  initI18n,
  logIn,
  logOut,
  assertOnlyVisibleToAuthenticated,
  assertOnlyVisibleToAdmin,
} from './shared';

initI18n();

// authenticated non-admin

test('Assert "Profile" button is visible upon logging in', async({ page }) => {
  await page.goto(process.env.FRONT_END_URL);
  await assertOnlyVisibleToAuthenticated(
    page, page.getByRole('button', { name: i18n.t('Profile') })
  );
});

// admin

test('Assert "Create form" button is (only) visible to admin', async({ page }) => {
  await page.goto(process.env.FRONT_END_URL);
  await assertOnlyVisibleToAdmin(
    page, page.getByRole('link', { name: i18n.t('navBarCreateForm')})
  );
});

test('Assert "Admin" button is (only) visible to admin', async({ page }) => {
  await page.goto(process.env.FRONT_END_URL);
  await assertOnlyVisibleToAdmin(
    page, page.getByRole('link', { name: i18n.t('navBarAdmin') })
  );
});
