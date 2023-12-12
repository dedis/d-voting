import { default as i18n } from 'i18next';
import { test, expect } from '@playwright/test';
import en from './../src/language/en.json';
import fr from './../src/language/fr.json';
import de from './../src/language/de.json';
import {
  SCIPER_ADMIN,
  SCIPER_USER,
  mockProxy,
  mockPersonalInfo,
  mockGetDevLogin,
  mockLogout,
} from './mocks';

export function initI18n () {
  i18n.init({
    resources: { en, fr, de },
    fallbackLng: ['en', 'fr', 'de'],
  });
}

export async function setUp(page: any, url: string) {
  await mockProxy(page);
  await mockGetDevLogin(page);
  await mockLogout(page);
  await page.goto(url);
  await expect(page).toHaveURL(url);   // make sure that page is loaded
}

export async function logIn (page: any, sciper: string) {
  await mockPersonalInfo(page, sciper);
  await page.reload();
  await expect(page).toHaveURL(page.url());   // make sure that page is loaded
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
  await logIn(page, SCIPER_ADMIN);
  await expect(locator).toBeVisible();    // assert is visible to admin user
}

export async function getFooter (page: any) {
  return await page.getByTestId('footer');
}
