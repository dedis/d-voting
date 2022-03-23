import { Select } from 'components/utils/types';
import { Answers, SELECT } from 'components/utils/useConfiguration';
import { t } from 'i18next';
import { getIndices } from './HandleAnswers';

const selectHintDisplay = (question: Select) => {
  let hint = '';
  let max = question.MaxN;
  let min = question.MinN;
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
  question: Select,
  choice: string,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) => {
  let { questionIndex, choiceIndex, errorIndex, newAnswers } = getIndices(
    question,
    choice,
    answers,
    SELECT
  );

  if (question.MaxN === 1) {
    newAnswers.SelectAnswers[questionIndex].Answers.fill(false);
  }
  newAnswers.SelectAnswers[questionIndex].Answers[choiceIndex] = e.target.checked;

  let numAnswer = newAnswers.SelectAnswers[questionIndex].Answers.filter((c) => c === true).length;
  if (numAnswer < question.MaxN + 1 && numAnswer >= question.MinN) {
    newAnswers.Errors[errorIndex].Message = '';
  } else if (numAnswer >= question.MaxN + 1) {
    newAnswers.Errors[errorIndex].Message =
      t('maxSelectError') + question.MaxN + t('pluralAnswers');
  }
  console.log('selectChecks: ' + JSON.stringify(answers));
  setAnswers(newAnswers);
};

const selectDisplay = (
  isChecked: boolean,
  choice: string,
  question: Select,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
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
        onChange={(e) => handleChecks(e, question, choice, answers, setAnswers)}
      />
      <label htmlFor={choice} className="pl-2 text-gray-700 cursor-pointer">
        {choice}
      </label>
    </div>
  );
};

export { selectDisplay, selectHintDisplay };
