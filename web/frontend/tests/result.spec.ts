import { expect, test } from '@playwright/test';
import { default as i18n } from 'i18next';
import { assertHasFooter, assertHasNavBar, initI18n, setUp } from './shared';
import { FORMID } from './mocks/shared';
import { mockFormsFormID } from './mocks/evoting';
import Form from './json/evoting/forms/result/clear.json';

initI18n();

test.beforeEach(async ({ page }) => {
  // TODO integrate localisation
  i18n.changeLanguage('en'); // force 'en' for these tests
  await mockFormsFormID(page, 5); // mock clear election result per default
  await setUp(page, `/forms/${FORMID}/result`);
});

test.afterAll(async () => {
  i18n.changeLanguage(); // unset language for the other tests
});

test('Assert navigation bar is present', async ({ page }) => {
  await assertHasNavBar(page);
});

test('Assert footer is present', async ({ page }) => {
  await assertHasFooter(page);
});

test('Assert form titles are displayed correctly', async ({ page }) => {
  await expect(page.getByText(i18n.t('navBarResult'))).toBeVisible();
  await expect(
    page.getByText(i18n.t('totalNumberOfVotes', { votes: Form.Result.length }))
  ).toBeVisible();
  const content = await page.getByTestId('content');
  await expect(content.locator('xpath=./div/div/div[2]/h3')).toContainText(
    Form.Configuration.Title.En
  );
  await expect(page.getByRole('tab', { name: i18n.t('resGroup') })).toBeVisible();
  await expect(page.getByRole('tab', { name: i18n.t('resIndiv') })).toBeVisible();
  await expect(content.locator('xpath=./div/div/div[2]/div/div[2]/div/div/div/h2')).toContainText(
    Form.Configuration.Scaffold.at(0).Title.En
  );
  await expect(
    content.locator('xpath=./div/div/div[2]/div/div[2]/div/div/div/div/div/div[1]/h2')
  ).toContainText(Form.Configuration.Scaffold.at(0).Selects.at(0).Title.En);
});

test('Assert grouped results are displayed correctly', async ({ page }) => {
  // grouped results are displayed by default
  const resultGrid = await page.getByTestId('content').locator('xpath=./div/div/div[2]/div/div[2]/div/div/div/div/div/div[2]');
  let j = 1;
  for (const [i, choice] of Form.Configuration.Scaffold.at(0).Selects.at(0).Choices.entries()) {
    await expect(resultGrid.locator(`xpath=./div[${j}]/span`)).toContainText(JSON.parse(choice).en);
    await expect(resultGrid.locator(`xpath=./div[${j+1}]/div/div[2]`)).toContainText(['33.33%', '33.33%', '66.67%'].at(i));
    j += 2;
  }
});

test('Assert individual results are displayed correctly', async ({ page }) => {
  // individual results are displayed after toggling view
  await page.getByRole('tab', { name: i18n.t('resIndiv') }).click();
  const content = await page.getByTestId('content');
  for (const i of [0, 1, 2]) {
    const result = [
      [true, false, false],
      [false, false, true],
      [false, true, true],
    ].at(i);
    for (const [j, choice] of Form.Configuration.Scaffold.at(0).Selects.at(0).Choices.entries()) {
      const resultRow = await content.locator(`xpath=./div/div/div[2]/div/div[2]/div/div/div[2]/div[1]/div[1]/div[2]/div[${j+1}]`);
      await expect(resultRow.locator('xpath=./div[2]')).toContainText(JSON.parse(choice).en);
      if (result.at(j)) {
        await expect(resultRow.getByRole('checkbox')).toBeChecked();
      } else {
        await expect(resultRow.getByRole('checkbox')).not.toBeChecked();
      }
    }
    if ([0, 1].includes(i)){
      await page.getByRole('button', { name: i18n.t('next') }).click();
    }
  }
});
