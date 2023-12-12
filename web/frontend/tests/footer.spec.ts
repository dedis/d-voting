import { expect, test } from '@playwright/test';
import { default as i18n } from 'i18next';
import { initI18n, setUp } from './shared';

initI18n();

test.beforeEach(async ({ page }) => {
  await setUp(page, '/about');
});

test('Assert copyright notice is visible', async ({ page }) => {
  const footerCopyright = await page.getByTestId('footerCopyright');
  await expect(footerCopyright).toBeVisible();
  await expect(footerCopyright).toHaveText(
    `© ${new Date().getFullYear()} ${i18n.t('footerCopyright')} https://github.com/c4dt/d-voting`
  );
});

test('Assert version information is visible', async ({ page }) => {
  const footerVersion = await page.getByTestId('footerVersion');
  await expect(footerVersion).toBeVisible();
  await expect(footerVersion).toHaveText(
    [
      `${i18n.t('footerVersion')} ${process.env.REACT_APP_VERSION || i18n.t('footerUnknown')}`,
      `${i18n.t('footerBuild')} ${process.env.REACT_APP_BUILD || i18n.t('footerUnknown')}`,
      `${i18n.t('footerBuildTime')} ${process.env.REACT_APP_BUILD_TIME || i18n.t('footerUnknown')}`,
    ].join(' - ')
  );
});
