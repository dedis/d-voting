import { expect, test } from '@playwright/test';
import { default as i18n } from 'i18next';
import { assertHasFooter, assertHasNavBar, initI18n, logIn, setUp } from './shared';
import { FORMID } from './mocks/shared';
import {
  SCIPER_ADMIN,
  SCIPER_OTHER_USER,
  SCIPER_USER,
  mockFormsVote,
  mockPersonalInfo,
} from './mocks/api';
import { mockFormsFormID } from './mocks/evoting';
import Form from './json/evoting/forms/open.json';

initI18n();

test.beforeEach(async ({ page }) => {
  await mockFormsFormID(page, 1);
  await logIn(page, SCIPER_ADMIN);
  await setUp(page, `/ballot/show/${FORMID}`);
});

test('Assert navigation bar is present', async ({ page }) => {
  await assertHasNavBar(page);
});

test('Assert footer is present', async ({ page }) => {
  await assertHasFooter(page);
});

test('Assert ballot form is correctly handled for anonymous users, non-voter users and voter users', async ({
  page,
}) => {
  const castVoteButton = await page.getByRole('button', { name: i18n.t('castVote') });
  await test.step('Assert anonymous is redirected to login page', async () => {
    await mockPersonalInfo(page);
    await page.reload({ waitUntil: 'networkidle' });
    await expect(page).toHaveURL('/login');
  });
  await test.step('Assert non-voter gets page that they are not allowed to vote', async () => {
    await logIn(page, SCIPER_OTHER_USER);
    await page.goto(`/ballot/show/${FORMID}`, { waitUntil: 'networkidle' });
    await expect(page).toHaveURL(`/ballot/show/${FORMID}`);
    await expect(castVoteButton).toBeHidden();
    await expect(page.getByText(i18n.t('voteNotVoter'))).toBeVisible();
    await expect(page.getByText(i18n.t('voteNotVoterDescription'))).toBeVisible();
  });
  await test.step('Assert voter gets ballot', async () => {
    await logIn(page, SCIPER_USER);
    await page.goto(`/ballot/show/${FORMID}`, { waitUntil: 'networkidle' });
    await expect(page).toHaveURL(`/ballot/show/${FORMID}`);
    await expect(castVoteButton).toBeVisible();
    await expect(page.getByText(i18n.t('vote'))).toBeVisible();
    await expect(page.getByText(i18n.t('voteExplanation'))).toBeVisible();
  });
});

test('Assert ballot is displayed properly', async ({ page }) => {
  const content = await page.getByTestId('content');
  // TODO integrate localisation
  i18n.changeLanguage('en'); // force 'en' for this test
  await expect(content.locator('xpath=./div/div[3]/h3')).toContainText(Form.Configuration.Title.En);
  const scaffold = Form.Configuration.Scaffold.at(0);
  await expect(content.locator('xpath=./div/div[3]/div/div/h3')).toContainText(scaffold.Title.En);
  const select = scaffold.Selects.at(0);
  await expect(
    content.locator('xpath=./div/div[3]/div/div/div/div/div/div[1]/div/h3')
  ).toContainText(select.Title.En);
  await expect(
    page.getByText(i18n.t('selectBetween', { minSelect: select.MinN, maxSelect: select.MaxN }))
  ).toBeVisible();
  for (const choice of select.Choices.map((x) => JSON.parse(x))) {
    await expect(page.getByRole('checkbox', { name: choice.en })).toBeVisible();
  }
  i18n.changeLanguage(); // unset language for the other tests
});

test('Assert minimum/maximum number of choices are handled correctly', async ({ page }) => {
  const castVoteButton = await page.getByRole('button', { name: i18n.t('castVote') });
  const select = Form.Configuration.Scaffold.at(0).Selects.at(0);
  await test.step(
    `Assert minimum number of choices (${select.MinN}) are handled correctly`,
    async () => {
      await castVoteButton.click();
      await expect(
        page.getByText(
          i18n.t('minSelectError', { min: select.MinN, singularPlural: i18n.t('singularAnswer') })
        )
      ).toBeVisible();
    }
  );
  await test.step(
    `Assert maximum number of choices (${select.MaxN}) are handled correctly`,
    async () => {
      for (const choice of select.Choices.map((x) => JSON.parse(x))) {
        await page.getByRole('checkbox', { name: choice.en }).setChecked(true);
      }
      await castVoteButton.click();
      await expect(page.getByText(i18n.t('maxSelectError', { max: select.MaxN }))).toBeVisible();
    }
  );
});

test('Assert that correct number of choices are accepted', async ({ page, baseURL }) => {
  await mockFormsVote(page);
  page.waitForRequest(async (request) => {
    const body = await request.postDataJSON();
    return (
      request.url() === `${baseURL}/api/evoting/forms/${FORMID}/vote` &&
      request.method() === 'POST' &&
      body.UserID === null &&
      body.Ballot.length === 1 &&
      body.Ballot.at(0).K.length === 32 &&
      body.Ballot.at(0).C.length === 32
    );
  });
  await page
    .getByRole('checkbox', {
      name: JSON.parse(Form.Configuration.Scaffold.at(0).Selects.at(0).Choices.at(0)).en,
    })
    .setChecked(true);
  await page.getByRole('button', { name: i18n.t('castVote') }).click();
});
