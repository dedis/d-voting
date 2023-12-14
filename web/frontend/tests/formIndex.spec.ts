import { expect, test } from '@playwright/test';
import { default as i18n } from 'i18next';
import { assertHasFooter, assertHasNavBar, initI18n, setUp } from './shared';
import { mockPersonalInfo, mockEvoting } from './mocks';

initI18n();

test.beforeEach(async ({ page }) => {
  // mock empty list per default
  await mockEvoting(page);
  await mockPersonalInfo(page);
  await setUp(page, '/form/index');
});

// main elements

test('Assert navigation bar is present', async ({ page }) => {
  await assertHasNavBar(page);
});

test('Assert footer is present', async ({ page }) => {
  await assertHasFooter(page);
});

// pagination bar

test('Assert pagination bar is present', async ({ page }) => {
  await expect(page.getByTestId('navPagination')).toBeVisible();
  await expect(page.getByRole('button', { name: i18n.t('previous') })).toBeVisible();
  await expect(page.getByRole('button', { name: i18n.t('next') })).toBeVisible();
});

test('Assert pagination works correctly for empty list', async ({ page }) => {
  await expect(page.getByTestId('navPaginationMessage')).toHaveText(i18n.t('showingNOverMOfXResults', { n: 1, m: 1, x: 0 }));
  for (let key of ['next', 'previous']) {
    await expect(page.getByRole('button', { name: i18n.t(key) })).toBeDisabled();
  }
});

test('Assert pagination works correctly for non-empty list', async ({ page }) => {
  // mock non-empty list w/ 11 elements i.e. 2 pages
  await mockEvoting(page, false);
  await page.reload();
  const next = await page.getByRole('button', { name: i18n.t('next') });
  const previous = await page.getByRole('button', { name: i18n.t('previous') });
  // 1st page
  await expect(page.getByTestId('navPaginationMessage')).toHaveText(i18n.t('showingNOverMOfXResults', { n: 1, m: 2, x: 11 }));
  await expect(previous).toBeDisabled();
  await expect(next).toBeEnabled();
  await next.click();
  // 2nd page
  await expect(page.getByTestId('navPaginationMessage')).toHaveText(i18n.t('showingNOverMOfXResults', { n: 2, m: 2, x: 11 }));
  await expect(next).toBeDisabled();
  await expect(previous).toBeEnabled();
  await previous.click();
  // back to 1st page
  await expect(page.getByTestId('navPaginationMessage')).toHaveText(i18n.t('showingNOverMOfXResults', { n: 1, m: 2, x: 11 }));
  await expect(previous).toBeDisabled();
  await expect(next).toBeEnabled();
});
