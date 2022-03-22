import { Error, TextAnswer } from 'components/utils/useConfiguration';
import { t } from 'i18next';
import { getIndexes } from './HandleAnswers';

const textHintDisplay = (question: any) => {
  let hint = '';
  let min = question.Content.MinN;
  let max = question.Content.MaxN;

  if (min !== max) {
    hint = t('minText') + min;
  } else {
    hint = t('fillText') + min;
  }
  hint += min > 1 ? t('pluralAnswers') : t('singularAnswer');
  return <div className="text-sm pl-2 pb-2 text-gray-400">{hint}</div>;
};

const handleTextInput = (
  e: React.ChangeEvent<HTMLInputElement>,
  question: any,
  choice: string,
  textStates: TextAnswer[],
  setTextStates: React.Dispatch<React.SetStateAction<TextAnswer[]>>,
  answerErrors: Error[],
  setAnswerErrors: React.Dispatch<React.SetStateAction<Error[]>>
) => {
  let { questionIndex, choiceIndex, errorIndex } = getIndexes(
    question,
    choice,
    textStates,
    answerErrors
  );
  let error = Array.from(answerErrors);
  let text = Array.from(textStates);
  text[questionIndex].Answers[choiceIndex] = e.target.value.trim();
  setTextStates(text);
  error[errorIndex].Message = '';

  if (question.Regex) {
    let regexp = new RegExp(question.Regex);
    for (const answer of text[questionIndex].Answers) {
      if (!regexp.test(answer) && answer !== '') {
        error[errorIndex].Message = t('regexpCheck') + question.Regex;
      }
    }
  }
  setAnswerErrors(error);
  console.log('textStates: ' + JSON.stringify(textStates));
};

const textDisplay = (
  choice: string,
  question: any,
  textStates: TextAnswer[],
  setTextStates: React.Dispatch<React.SetStateAction<TextAnswer[]>>,
  answerErrors: Error[],
  setAnswerErrors: React.Dispatch<React.SetStateAction<Error[]>>
) => {
  return (
    <div>
      <label htmlFor={choice}>{choice + ': '}</label>
      <input
        id={choice}
        type="text"
        key={choice}
        className="mt-1 sm:text-sm border rounded-md w-1/2"
        onChange={(e) =>
          handleTextInput(
            e,
            question,
            choice,
            textStates,
            setTextStates,
            answerErrors,
            setAnswerErrors
          )
        }
      />
    </div>
  );
};

export { textDisplay, textHintDisplay };
