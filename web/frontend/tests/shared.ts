import { default as i18n } from 'i18next';
import en from './../src/language/en.json';
import fr from './../src/language/fr.json';
import de from './../src/language/de.json';
import { SCIPER_ADMIN, SCIPER_USER, mockGetDevLogin, mockLogout } from './mocks';

export function initI18n () {
  i18n.init({
    resources: { en, fr, de },
    fallbackLng: ['en', 'fr', 'de'],
  });
}

export async function logIn (page: any, admin = false) {
  await mockGetDevLogin(page);
  await page.context().request.get(
    `${process.env.FRONT_END_URL}/api/get_dev_login/${admin ? SCIPER_ADMIN : SCIPER_USER}`
  );
  await page.reload();
}

export async function logOut (page: any) {
  await mockLogout(page);
  await page.context().request.post(`${process.env.FRONT_END_URL}/api/logout`);
  await page.reload();
}
