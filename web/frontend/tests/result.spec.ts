import { expect, test } from '@playwright/test';
import { default as i18n } from 'i18next';
import { assertHasFooter, assertHasNavBar, initI18n, setUp } from './shared';
import { FORMID } from './mocks/shared';
import { mockFormsFormID } from './mocks/evoting';
import Form from './json/evoting/forms/combined.json';

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
  for (const [index, scaffold] of Form.Configuration.Scaffold.entries()) {
    await expect(
      content.locator(`xpath=./div/div/div[2]/div/div[2]/div/div/div[${index + 1}]/h2`)
    ).toContainText(scaffold.Title.En);
    await expect(
      content.locator(
        `xpath=./div/div/div[2]/div/div[2]/div/div/div[${index + 1}]/div/div/div[1]/h2`
      )
    ).toContainText(scaffold.Selects.at(0).Title.En);
  }
});

test('Assert grouped results are displayed correctly', async ({ page }) => {
  // grouped results are displayed by default
  let i = 1;
  for (const expected of [
    [
      ['Blue', '3/4'],
      ['Green', '2/4'],
      ['Red', '1/4'],
    ],
    [
      ['Cyan', '2/4'],
      ['Magenta', '2/4'],
      ['Yellow', '1/4'],
      ['Key', '1/4'],
    ],
  ]) {
    const resultGrid = await page
      .getByTestId('content')
      .locator(`xpath=./div/div/div[2]/div/div[2]/div/div/div[${i}]/div/div/div[2]`);
    i += 1;
    let j = 1;
    for (const [title, totalCount] of expected) {
      await expect(resultGrid.locator(`xpath=./div[${j}]/span`)).toContainText(title);
      await expect(resultGrid.locator(`xpath=./div[${j + 1}]/div/div[2]`)).toContainText(
        totalCount
      );
      j += 2;
    }
  }
});

test('Assert individual results are displayed correctly', async ({ page }) => {
  // individual results are displayed after toggling view
  await page.getByRole('tab', { name: i18n.t('resIndiv') }).click();
  const content = await page.getByTestId('content');
  for (const i of [0, 1, 2, 3]) {
    // for each ballot
    for (const [index, scaffold] of Form.Configuration.Scaffold.entries()) {
      // for each form
      const result = [
        [
          [true, false, false],
          [true, true, false, false],
        ],
        [
          [false, true, true],
          [true, false, false, true],
        ],
        [
          [false, true, true],
          [false, true, false, false],
        ],
        [
          [false, false, true],
          [false, false, true, false],
        ],
      ]
        .at(i)
        .at(index); // get results for this ballot and this form
      for (const [j, choice] of scaffold.Selects.at(0).Choices.entries()) {
        const resultRow = await content.locator(
          `xpath=./div/div/div[2]/div/div[2]/div/div/div[${index + 2}]/div[1]/div[1]/div[2]/div[${
            j + 1
          }]`
        );
        await expect(resultRow.locator('xpath=./div[2]')).toContainText(JSON.parse(choice).en);
        if (result.at(j)) {
          await expect(resultRow.getByRole('checkbox')).toBeChecked();
        } else {
          await expect(resultRow.getByRole('checkbox')).not.toBeChecked();
        }
      }
    }
    if ([0, 1, 2].includes(i)) {
      // display next ballot
      await page.getByRole('button', { name: i18n.t('next') }).click();
    }
  }
});
