import { Error, Question, SelectAnswer, TextAnswer } from 'components/utils/useConfiguration';
import { t } from 'i18next';

const getIndexes = (question: any, choice: string, list: any, answerErrors: Error[]) => {
  let questionIndex = list.findIndex((s: any) => s.ID === question.ID);
  let choiceIndex = question.Choices.findIndex((c: string) => c === choice);
  let errorIndex = answerErrors.findIndex((error) => error.ID === question.ID);
  return { questionIndex, choiceIndex, errorIndex };
};

const ballotIsValid = (
  sortedQuestion: Question[],
  selectStates: SelectAnswer[],
  textStates: TextAnswer[],
  answerErrors: Error[],
  setAnswerErrors: React.Dispatch<React.SetStateAction<Error[]>>
) => {
  let isValid = true;
  let errors = Array.from(answerErrors);

  for (const select of selectStates) {
    let numAnswer = select.Answers.filter((answer) => answer === true).length;
    let selectQuestion = sortedQuestion.find((s: any) => s.Content.ID === select.ID);
    let errorIndex = answerErrors.findIndex((e) => e.ID === select.ID);
    if (numAnswer < selectQuestion.Content.MinN) {
      errors[errorIndex].Message = t('minSelectError') + selectQuestion.Content.MinN;
      errors[errorIndex].Message +=
        selectQuestion.Content.MinN > 1 ? t('pluralAnswers') : t('singularAnswer');
      setAnswerErrors(errors);
      isValid = false;
    }

    if (numAnswer > selectQuestion.Content.MaxN) {
      isValid = false;
    }
  }

  for (const text of textStates) {
    let textQuestion = sortedQuestion.find((s: any) => s.Content.ID === text.ID);
    if (textQuestion.Content.Regex) {
      let regexp = new RegExp(textQuestion.Content.Regex);
      for (const answer of text.Answers) {
        if (!regexp.test(answer) && answer !== '') {
          isValid = false;
        }
      }
    }

    let numAnswer = text.Answers.filter((answer) => answer !== '').length;
    let errorIndex = answerErrors.findIndex((e) => e.ID === text.ID);
    if (numAnswer < textQuestion.Content.MinN) {
      errors[errorIndex].Message = t('minTextError') + textQuestion.Content.MinN;
      errors[errorIndex].Message +=
        textQuestion.Content.MinN > 1 ? t('pluralAnswers') : t('singularAnswer');
      setAnswerErrors(errors);
      isValid = false;
    }
  }

  return isValid;
};

export { ballotIsValid, getIndexes };
