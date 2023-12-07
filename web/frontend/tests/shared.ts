import { default as i18n } from 'i18next';
import { test, expect } from '@playwright/test';
import en from './../src/language/en.json';
import fr from './../src/language/fr.json';
import de from './../src/language/de.json';
import { mockPersonalInfo, mockGetDevLogin, mockLogout } from './mocks';

export function initI18n () {
  i18n.init({
    resources: { en, fr, de },
    fallbackLng: ['en', 'fr', 'de'],
  });
}

export const SCIPER_ADMIN = '123456';
export const SCIPER_USER = '789012';

export async function logIn (page: any, sciper) {
  await mockGetDevLogin(page, sciper);
  await mockPersonalInfo(page, sciper);
  // uncomment the following line to update the HAR files
  await page.context().request.get(`${process.env.FRONT_END_URL}/api/get_dev_login/${sciper}`);
  await page.reload();
  await expect(page).toHaveURL(page.url());   // make sure that page is loaded correctly
}

export async function logOut (page: any, sciper) {
  await mockLogout(page, sciper);
//  await mockPersonalInfo(page, null);
  // uncomment the following line to update the HAR files
  await page.context().request.post(`${process.env.FRONT_END_URL}/api/logout`);
  await page.reload();
  await expect(page).toHaveURL(page.url());   // make sure that page is loaded correctly
}

export async function assertOnlyVisibleToAuthenticated (page: any, locator: any) {
  await expect(locator).toBeHidden();   // assert is hidden to unauthenticated user
  await logIn(page, SCIPER_USER);
  await expect(locator).toBeVisible();  // assert is visible to authenticated user
}

export async function assertOnlyVisibleToAdmin (page: any, locator: any) {
  await expect(locator).toBeHidden();     // assert is hidden to unauthenticated user
  await logIn(page, SCIPER_USER);
  await expect(locator).toBeHidden();     // assert is hidden to authenticated non-admin user
  await logOut(page);
  await logIn(page, SCIPER_ADMIN);
  await expect(locator).toBeVisible();    // assert is visible to admin user
}
