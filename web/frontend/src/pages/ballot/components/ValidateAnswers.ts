import { t } from 'i18next';
import { RANK, SELECT, SUBJECT, TEXT } from 'types/configuration';
import * as types from 'types/configuration';
import { answersFrom } from 'types/getObjectType';

function selectAnswerIsValid(selectQuestion: types.SelectQuestion, newAnswers: types.Answers) {
  const numAnswer = newAnswers.SelectAnswers.get(selectQuestion.ID).filter(
    (answer) => answer === true
  ).length;

  let isValid = true;
  let selectError = newAnswers.Errors.get(selectQuestion.ID);

  if (numAnswer < selectQuestion.MinN) {
    selectError =
      selectQuestion.MinN > 1
        ? t('minSelectError', { min: selectQuestion.MinN, singularPlural: t('pluralAnswers') })
        : t('minSelectError', { min: selectQuestion.MinN, singularPlural: t('singularAnswer') });

    isValid = false;
  }

  if (numAnswer > selectQuestion.MaxN) {
    isValid = false;
  }

  newAnswers.Errors.set(selectQuestion.ID, selectError);

  return isValid;
}

function textAnswerIsValid(textQuestion: types.TextQuestion, newAnswers: types.Answers) {
  const textAnswer = newAnswers.TextAnswers.get(textQuestion.ID);
  const numAnswer = textAnswer.filter((answer) => answer !== '').length;
  let textError = newAnswers.Errors.get(textQuestion.ID);
  let isValid = true;

  for (const answer of textAnswer) {
    if (answer.length > textQuestion.MaxLength) {
      textError = t('maxTextChars', {
        maxLength: textQuestion.MaxLength,
      });

      isValid = false;
    }

    let regexp = new RegExp(textQuestion.Regex);

    if (!regexp.test(answer) && answer !== '') {
      textError = t('regexpCheck', { regexp: textQuestion.Regex });
      isValid = false;
    }
  }

  if (numAnswer < textQuestion.MinN) {
    textError =
      textQuestion.MinN > 1
        ? t('minTextError', { minText: textQuestion.MinN, singularPlural: t('pluralAnswers') })
        : t('minTextError', { minText: textQuestion.MinN, singularPlural: t('singularAnswer') });

    isValid = false;
  }

  newAnswers.Errors.set(textQuestion.ID, textError);

  return isValid;
}

function subjectIsValid(subject: types.Subject, newAnswers: types.Answers) {
  let elementIsValid = true;
  let isValid = true;

  subject.Elements.forEach((element) => {
    switch (element.Type) {
      case RANK:
        // TODO: when implementing the new ranks
        break;
      case SELECT:
        elementIsValid = selectAnswerIsValid(element as types.SelectQuestion, newAnswers);
        break;
      case TEXT:
        elementIsValid = textAnswerIsValid(element as types.TextQuestion, newAnswers);
        break;
      case SUBJECT:
        elementIsValid = subjectIsValid(element as types.Subject, newAnswers);
    }
    isValid = isValid && elementIsValid;
  });

  return isValid;
}

export function ballotIsValid(
  configuration: types.Configuration,
  answers: types.Answers,
  setAnswers: React.Dispatch<React.SetStateAction<types.Answers>>
) {
  let isValid = true;
  let newAnswers = answersFrom(answers);
  let subjIsValid = true;
  for (const subject of configuration.Scaffold) {
    subjIsValid = subjectIsValid(subject, newAnswers);
    isValid = isValid && subjIsValid;
  }
  setAnswers(newAnswers);

  return isValid;
}
