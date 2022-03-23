import { Rank, Select, Text } from 'components/utils/types';
import {
  Answers,
  Error,
  Question,
  RANK,
  RankAnswer,
  SELECT,
  SelectAnswer,
  TEXT,
  TextAnswer,
} from 'components/utils/useConfiguration';
import { t } from 'i18next';

const buildAnswer = (answers: Answers) => {
  let newAnswers: Answers = {
    SelectAnswers: Array.from(answers.SelectAnswers),
    RankAnswers: Array.from(answers.RankAnswers),
    TextAnswers: Array.from(answers.TextAnswers),
    Errors: Array.from(answers.Errors),
  };

  return newAnswers;
};

const getIndices = (
  question: Select | Rank | Text,
  choice: string,
  answers: Answers,
  type: string
) => {
  let questionIndex: number;
  switch (type) {
    case RANK:
      questionIndex = answers.RankAnswers.findIndex((r: RankAnswer) => r.ID === question.ID);
      break;
    case SELECT:
      questionIndex = answers.SelectAnswers.findIndex((s: SelectAnswer) => s.ID === question.ID);
      break;
    case TEXT:
      questionIndex = answers.TextAnswers.findIndex((t: TextAnswer) => t.ID === question.ID);
  }
  let choiceIndex = question.Choices.findIndex((c: string) => c === choice);
  let errorIndex = answers.Errors.findIndex((e: Error) => e.ID === question.ID);

  let newAnswers: Answers = buildAnswer(answers);

  return { questionIndex, choiceIndex, errorIndex, newAnswers };
};

const ballotIsValid = (sortedQuestion: Question[], answers, setAnswers) => {
  let isValid = true;
  let newAnswers = buildAnswer(answers);

  for (const select of answers.SelectAnswers) {
    let numAnswer = select.Answers.filter((answer) => answer === true).length;
    let selectQuestion = sortedQuestion.find((s) => s.Content.ID === select.ID).Content as Select;
    let errorIndex = newAnswers.Errors.findIndex((e) => e.ID === select.ID);
    if (numAnswer < selectQuestion.MinN) {
      newAnswers.Errors[errorIndex].Message = t('minSelectError') + selectQuestion.MinN;
      newAnswers.Errors[errorIndex].Message +=
        selectQuestion.MinN > 1 ? t('pluralAnswers') : t('singularAnswer');
      isValid = false;
    }

    if (numAnswer > selectQuestion.MaxN) {
      isValid = false;
    }
  }

  for (const text of answers.TextAnswers) {
    let textQuestion = sortedQuestion.find((s: any) => s.Content.ID === text.ID).Content as Text;
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
      newAnswers.Errors[errorIndex].Message = t('minTextError') + textQuestion.MinN;
      newAnswers.Errors[errorIndex].Message +=
        textQuestion.MinN > 1 ? t('pluralAnswers') : t('singularAnswer');
      isValid = false;
    }
  }
  setAnswers(newAnswers);
  return isValid;
};

export { ballotIsValid, getIndices, buildAnswer };
