import { expect, test } from '@playwright/test';
import { default as i18n } from 'i18next';
import { assertHasFooter, assertHasNavBar, initI18n, setUp, translate, logIn } from './shared';
import { mockEvoting, mockPersonalInfo, SCIPER_USER, SCIPER_ADMIN } from './mocks';
import Forms from './json/formList.json';
import User from './json/personal_info/789012.json';
import Admin from './json/personal_info/123456.json';

initI18n();

async function goForward(page: page) {
  await page.getByRole('button', { name: i18n.t('next') }).click();
}

async function goBackward(page: page) {
  await page.getByRole('button', { name: i18n.t('previous') }).click();
}

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
  await expect(page.getByTestId('navPaginationMessage')).toHaveText(
    i18n.t('showingNOverMOfXResults', { n: 1, m: 1, x: 0 })
  );
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
  await expect(page.getByTestId('navPaginationMessage')).toHaveText(
    i18n.t('showingNOverMOfXResults', { n: 1, m: 2, x: 11 })
  );
  await expect(previous).toBeDisabled();
  await expect(next).toBeEnabled();
  await next.click();
  // 2nd page
  await expect(page.getByTestId('navPaginationMessage')).toHaveText(
    i18n.t('showingNOverMOfXResults', { n: 2, m: 2, x: 11 })
  );
  await expect(next).toBeDisabled();
  await expect(previous).toBeEnabled();
  await previous.click();
  // back to 1st page
  await expect(page.getByTestId('navPaginationMessage')).toHaveText(
    i18n.t('showingNOverMOfXResults', { n: 1, m: 2, x: 11 })
  );
  await expect(previous).toBeDisabled();
  await expect(next).toBeEnabled();
});


test('Assert no forms are displayed for empty list', async ({ page }) => {
  // 1 header row
  await expect.poll(async () => {
    const rows = await page.getByRole('table').getByRole('row');
    return rows.all();
  }).toHaveLength(1);
});

async function assertQuickAction(row: locator, form: object, sciper?: string) {
  const user = sciper === SCIPER_USER ? User
          : sciper === SCIPER_ADMIN ? Admin
          : undefined;
  const quickAction = row.getByTestId('quickAction');
  switch(form.Status) {
    case 1:
      // only authenticated user w/ right to vote sees 'vote' button
      if ((user) && (form.FormID in user.authorization) && (user.authorization[form.FormID].includes('vote'))) {
        await expect(quickAction).toHaveText(i18n.t('vote'));
        await expect(await quickAction.getByRole('link')).toHaveAttribute('href', `/ballot/show/${form.FormID}`);
        await expect(quickAction).toBeVisible();
      }
      else {
        await expect(quickAction).toBeHidden();
      }
      break;
    case 5:
      // any user can see the results of a past election
      await expect(quickAction).toHaveText(i18n.t('seeResult'));
      await expect(await quickAction.getByRole('link')).toHaveAttribute('href', `/forms/${form.FormID}/result`);
      break;
     default:
      await expect(quickAction).toBeHidden();
  }
}

test('Assert forms are displayed correctly for unauthenticated user', async ({ page, baseURL }) => {
  await mockEvoting(page, false);
  await page.reload();
  const table = await page.getByRole('table');
  for (let form of Forms.Forms.slice(0, -1)) {
    let name = translate(form.Title);
    let row = await table.getByRole('row', { name: name });
    await expect(row).toBeVisible();
    // row entry leads to form view
    let link = await row.getByRole('link', { name: name });
    await expect(link).toBeVisible();
    await expect(link).toHaveAttribute('href', `/forms/${form.FormID}`);
    await assertQuickAction(row, form);
  }
  await goForward(page);
  let row = await table.getByRole('row', { name: translate(Forms.Forms.at(-1).Title) });
  await expect(row).toBeVisible();
  await assertQuickAction(row, Forms.Forms.at(-1));
});

test('Assert quick actions are displayed correctly for authenticated users', async ({ page, baseURL }) => {
  for (let sciper of [SCIPER_USER, SCIPER_ADMIN]) {
    await logIn(page, sciper);
    await mockEvoting(page, false);
    await page.reload();
    const table = await page.getByRole('table');
    for (let form of Forms.Forms.slice(0, -1)) {
      let row = await table.getByRole('row', { name: translate(form.Title) });
      await assertQuickAction(row, form, sciper);
    }
    await goForward(page);
    await assertQuickAction(await table.getByRole('row', { name: translate(Forms.Forms.at(-1).Title) }), Forms.Forms.at(-1));
  }
});
