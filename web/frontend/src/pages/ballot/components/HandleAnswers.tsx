import { t } from 'i18next';
import {
  Answers,
  Error,
  Question,
  RANK,
  RankAnswer,
  RankQuestion,
  SELECT,
  SelectAnswer,
  SelectQuestion,
  TEXT,
  TextAnswer,
  TextQuestion,
} from 'types/configuration';

export function buildAnswer(answers: Answers) {
  let newAnswers: Answers = {
    SelectAnswers: Array.from(answers.SelectAnswers),
    RankAnswers: Array.from(answers.RankAnswers),
    TextAnswers: Array.from(answers.TextAnswers),
    Errors: Array.from(answers.Errors),
  };

  return newAnswers;
}

export function getIndices(
  question: SelectQuestion | RankQuestion | TextQuestion,
  choice: string,
  answers: Answers,
  type: string
) {
  let questionIndex: number;

  switch (type) {
    case RANK:
      questionIndex = answers.RankAnswers.findIndex((rank: RankAnswer) => rank.ID === question.ID);
      break;
    case SELECT:
      questionIndex = answers.SelectAnswers.findIndex(
        (select: SelectAnswer) => select.ID === question.ID
      );
      break;
    case TEXT:
      questionIndex = answers.TextAnswers.findIndex((text: TextAnswer) => text.ID === question.ID);
  }

  let choiceIndex = question.Choices.findIndex((c: string) => c === choice);
  let errorIndex = answers.Errors.findIndex((e: Error) => e.ID === question.ID);

  let newAnswers: Answers = buildAnswer(answers);

  return { questionIndex, choiceIndex, errorIndex, newAnswers };
}

export function ballotIsValid(
  sortedQuestion: Question[],
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) {
  let isValid = true;
  let newAnswers = buildAnswer(answers);

  for (const select of answers.SelectAnswers) {
    let numAnswer = select.Answers.filter((answer) => answer === true).length;
    let selectQuestion = sortedQuestion.find((s) => s.Content.ID === select.ID)
      .Content as SelectQuestion;
    let errorIndex = newAnswers.Errors.findIndex((e) => e.ID === select.ID);

    if (numAnswer < selectQuestion.MinN) {
      newAnswers.Errors[errorIndex].Message =
        selectQuestion.MinN > 1
          ? t('minSelectError', { min: selectQuestion.MinN, singularPlural: t('pluralAnswers') })
          : t('minSelectError', { min: selectQuestion.MinN, singularPlural: t('singularAnswer') });
      isValid = false;
    }

    if (numAnswer > selectQuestion.MaxN) {
      isValid = false;
    }
  }

  for (const text of answers.TextAnswers) {
    let textQuestion = sortedQuestion.find((s: any) => s.Content.ID === text.ID)
      .Content as TextQuestion;

    if (textQuestion.Regex) {
      let regexp = new RegExp(textQuestion.Regex);
      for (const answer of text.Answers) {
        if (!regexp.test(answer) && answer !== '') {
          isValid = false;
        }
      }
    }

    let numAnswer = text.Answers.filter((answer) => answer !== '').length;
    let errorIndex = newAnswers.Errors.findIndex((e) => e.ID === text.ID);

    if (numAnswer < textQuestion.MinN) {
      newAnswers.Errors[errorIndex].Message =
        textQuestion.MinN > 1
          ? t('minTextError', { minText: textQuestion.MinN, singularPlural: t('pluralAnswers') })
          : t('minTextError', { minText: textQuestion.MinN, singularPlural: t('singularAnswer') });
      isValid = false;
    }
  }
  setAnswers(newAnswers);

  return isValid;
}
