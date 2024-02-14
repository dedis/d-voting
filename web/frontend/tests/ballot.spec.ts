import { expect, test } from '@playwright/test';
import { default as i18n } from 'i18next';
import { assertHasFooter, assertHasNavBar, initI18n, logIn, setUp } from './shared';
import { FORMID } from './mocks/shared';
import { SCIPER_ADMIN, SCIPER_OTHER_USER, SCIPER_USER, mockPersonalInfo } from './mocks/api';
import { mockFormsFormID } from './mocks/evoting';

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
