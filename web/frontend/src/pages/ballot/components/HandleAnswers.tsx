import { t } from 'i18next';
import {
  Answers,
  Configuration,
  RANK,
  SELECT,
  SUBJECT,
  SelectQuestion,
  Subject,
  TEXT,
  TextQuestion,
} from 'types/configuration';
import { answersFrom } from 'types/getObjectType';

function isSelectAnswerValid(
  selectQuestion: SelectQuestion,
  isValid: boolean,
  newAnswers: Answers
) {
  const numAnswer = newAnswers.SelectAnswers.get(selectQuestion.ID).filter(
    (answer) => answer === true
  ).length;

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

function isTextAnswerValid(textQuestion: TextQuestion, isValid: boolean, newAnswers: Answers) {
  const textAnswer = newAnswers.TextAnswers.get(textQuestion.ID);
  const numAnswer = textAnswer.filter((answer) => answer !== '').length;
  let textError = newAnswers.Errors.get(textQuestion.ID);

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

function isSubjectValid(subject: Subject, isValid: boolean, newAnswers: Answers) {
  let elementIsValid = true;
  for (const element of Array.from(subject.Elements.values())) {
    switch (element.Type) {
      case RANK:
        // TODO: when implementing the new ranks
        break;
      case SELECT:
        elementIsValid = isSelectAnswerValid(element as SelectQuestion, isValid, newAnswers);
        break;
      case TEXT:
        elementIsValid = isTextAnswerValid(element as TextQuestion, isValid, newAnswers);
        break;
      case SUBJECT:
        elementIsValid = isSubjectValid(element as Subject, isValid, newAnswers);
    }
    isValid = isValid && elementIsValid;
  }
  return isValid;
}

export function ballotIsValid(
  configuration: Configuration,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) {
  let isValid = true;
  let newAnswers = answersFrom(answers);
  let subjectIsValid = true;
  for (const subject of configuration.Scaffold) {
    subjectIsValid = isSubjectValid(subject as Subject, isValid, newAnswers);
    isValid = isValid && subjectIsValid;
  }
  setAnswers(newAnswers);

  return isValid;
}
