import { expect, test } from '@playwright/test';
import { default as i18n } from 'i18next';
import { assertHasFooter, assertHasNavBar, initI18n, setUp } from './shared';
import { FORMID } from './mocks/shared';
import { mockFormsFormID } from './mocks/evoting';
import Form from './json/evoting/forms/result/clear.json';

initI18n();

test.beforeEach(async ({ page }) => {
  await mockFormsFormID(page, 5); // mock clear election result per default
  await setUp(page, `/forms/${FORMID}/result`);
});

test('Assert navigation bar is present', async ({ page }) => {
  await assertHasNavBar(page);
});

test('Assert footer is present', async ({ page }) => {
  await assertHasFooter(page);
});

test('Assert form is displayed correctly', async ({ page }) => {
  // TODO integrate localisation
  i18n.changeLanguage('en'); // force 'en' for this test
  await expect(page.getByText(i18n.t('navBarResult'))).toBeVisible();
  await expect(
    page.getByText(i18n.t('totalNumberOfVotes', { votes: Form.Result.length }))
  ).toBeVisible();
  const content = await page.getByTestId('content');
  await expect(content.locator('xpath=./div/div/div[2]/h3')).toContainText(
    Form.Configuration.Title.En
  );
  await expect(content.locator('xpath=./div/div/div[2]/div/div[2]/div/div/div/h2')).toContainText(
    Form.Configuration.Scaffold.at(0).Title.En
  );
  await expect(
    content.locator('xpath=./div/div/div[2]/div/div[2]/div/div/div/div/div/div[1]/h2')
  ).toContainText(Form.Configuration.Scaffold.at(0).Selects.at(0).Title.En);
  for (const choice of Form.Configuration.Scaffold.at(0).Selects.at(0).Choices) {
    await expect(page.getByText(JSON.parse(choice).en)).toBeVisible();
  }
  i18n.changeLanguage(); // unset language for the other tests
});
