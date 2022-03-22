import { Error, SelectAnswer } from 'components/utils/useConfiguration';
import { t } from 'i18next';
import { getIndexes } from './HandleAnswers';

const selectHintDisplay = (question: any) => {
  let hint = '';
  let max = question.Content.MaxN;
  let min = question.Content.MinN;
  if (max === min) {
    hint = t('select') + min;
    hint += max > 1 ? t('pluralAnswers') : t('singularAnswer');
  } else {
    hint = t('selectBetween') + min + t('selectAnd') + max + t('pluralAnswers');
  }

  return <div className="text-sm pl-2 pb-2 text-gray-400">{hint}</div>;
};

const handleChecks = (
  e: React.ChangeEvent<HTMLInputElement>,
  question: any,
  choice: string,
  selectStates: SelectAnswer[],
  setSelectStates: React.Dispatch<React.SetStateAction<SelectAnswer[]>>,
  answerErrors: Error[],
  setAnswerErrors: React.Dispatch<React.SetStateAction<Error[]>>
) => {
  let { questionIndex, choiceIndex, errorIndex } = getIndexes(
    question,
    choice,
    selectStates,
    answerErrors
  );
  let error = Array.from(answerErrors);
  let newAnswer = Array.from(selectStates);
  if (question.MaxN === 1) {
    newAnswer[questionIndex].Answers.fill(false);
  }
  newAnswer[questionIndex].Answers[choiceIndex] = e.target.checked;
  setSelectStates(newAnswer);

  let numAnswer = selectStates[questionIndex].Answers.filter((c) => c === true).length;
  if (numAnswer < question.MaxN + 1 && numAnswer >= question.MinN) {
    error[errorIndex].Message = '';
  } else if (numAnswer >= question.MaxN + 1) {
    error[errorIndex].Message = t('maxSelectError') + question.MaxN + t('pluralAnswers');
  }
  console.log('selectChecks: ' + JSON.stringify(selectStates));
  setAnswerErrors(error);
};

const selectDisplay = (
  isChecked: boolean,
  choice: string,
  question: any,
  selectStates: SelectAnswer[],
  setSelectStates: React.Dispatch<React.SetStateAction<SelectAnswer[]>>,
  answerErrors: Error[],
  setAnswerErrors: React.Dispatch<React.SetStateAction<Error[]>>
) => {
  return (
    <div>
      <input
        id={choice}
        className="h-4 w-4 mt-1 mr-2 cursor-pointer accent-indigo-500"
        type="checkbox"
        key={choice}
        value={choice}
        checked={isChecked}
        onChange={(e) =>
          handleChecks(
            e,
            question,
            choice,
            selectStates,
            setSelectStates,
            answerErrors,
            setAnswerErrors
          )
        }
      />
      <label htmlFor={choice} className="pl-2 text-gray-700 cursor-pointer">
        {choice}
      </label>
    </div>
  );
};

export { selectDisplay, selectHintDisplay };
