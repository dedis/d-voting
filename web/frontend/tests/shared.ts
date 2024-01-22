import { default as i18n } from 'i18next';
import { expect } from '@playwright/test';
import en from './../src/language/en.json';
import fr from './../src/language/fr.json';
import de from './../src/language/de.json';
import {
  SCIPER_ADMIN,
  SCIPER_USER,
  mockGetDevLogin,
  mockLogout,
  mockPersonalInfo,
  mockProxy,
} from './mocks/api';

export function initI18n() {
  i18n.init({
    resources: { en, fr, de },
    fallbackLng: ['en', 'fr', 'de'],
  });
}

export async function setUp(page: page, url: string) {
  await mockProxy(page);
  await mockGetDevLogin(page);
  await mockLogout(page);
  // make sure that page is loaded
  await page.goto(url, { waitUntil: 'networkidle' });
  await expect(page).toHaveURL(url);
}

export async function logIn(page: page, sciper: string) {
  await mockPersonalInfo(page, sciper);
  await page.reload({ waitUntil: 'networkidle' });
}

export async function assertOnlyVisibleToAuthenticated(page: page, locator: locator) {
  await expect(locator).toBeHidden(); // assert is hidden to unauthenticated user
  await logIn(page, SCIPER_USER);
  await expect(locator).toBeVisible(); // assert is visible to authenticated user
}

export async function assertOnlyVisibleToAdmin(page: page, locator: locator) {
  await expect(locator).toBeHidden(); // assert is hidden to unauthenticated user
  await logIn(page, SCIPER_USER);
  await expect(locator).toBeHidden(); // assert is hidden to authenticated non-admin user
  await logIn(page, SCIPER_ADMIN);
  await expect(locator).toBeVisible(); // assert is visible to admin user
}

export async function assertHasNavBar(page: page) {
  await expect(page.getByTestId('navBar')).toBeVisible();
}

export async function assertHasFooter(page: page) {
  await expect(page.getByTestId('footer')).toBeVisible();
}

export function translate(internationalizable: object) {
  switch (i18n.language) {
    case 'en':
      return internationalizable.En;
    case 'fr':
      return internationalizable.Fr || internationalizable.En;
    case 'de':
      return internationalizable.De || internationalizable.En;
    default:
      return internationalizable.En;
  }
}
