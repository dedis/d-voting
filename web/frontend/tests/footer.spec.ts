import { test, expect } from '@playwright/test';
import { default as i18n } from 'i18next';
import { initI18n, setUp, getFooter } from './shared';

initI18n();

test.beforeEach(async ({ page }) => {
  await setUp(page, '/about');
});

test('Assert copyright notice is visible', async({ page }) => {
  const footerCopyright = await page.getByTestId('footerCopyright');
  await expect(footerCopyright).toBeVisible();
  await expect(footerCopyright).toHaveText(
    `© ${new Date().getFullYear()} ${i18n.t('footerCopyright')} https://github.com/c4dt/d-voting`
  );
});
