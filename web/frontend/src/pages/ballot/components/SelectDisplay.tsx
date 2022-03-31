import { FC } from 'react';
import { useTranslation } from 'react-i18next';
import { Answers, SELECT, SelectQuestion } from 'types/configuration';
import { getIndices } from './HandleAnswers';
import { HintDisplayProps } from '../QuestionDisplay';

const SelectHintDisplay: FC<HintDisplayProps> = ({ questionContent }) => {
  const content = questionContent as SelectQuestion;
  const { t } = useTranslation();

  let hint = '';
  let max = content.MaxN;
  let min = content.MinN;
  if (max === min) {
    hint =
      max > 1
        ? t('select', { minSelect: min, singularPlural: t('pluralAnswers') })
        : t('select', { minSelect: min, singularPlural: t('singularAnswer') });
  } else {
    hint = t('selectBetween', { minSelect: min, maxSelect: max });
  }

  return <div className="text-sm pl-2 pb-2 text-gray-400">{hint}</div>;
};

type SelectDisplayProps = {
  isChecked: boolean;
  choice: string;
  question: SelectQuestion;
  answers: Answers;
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>;
};

const SelectDisplay: FC<SelectDisplayProps> = ({
  isChecked,
  choice,
  question,
  answers,
  setAnswers,
}) => {
  const { t } = useTranslation();

  const handleChecks = (e: React.ChangeEvent<HTMLInputElement>) => {
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
    let numAnswer = newAnswers.SelectAnswers[questionIndex].Answers.filter(
      (c) => c === true
    ).length;

    if (numAnswer < question.MaxN + 1 && numAnswer >= question.MinN) {
      newAnswers.Errors[errorIndex].Message = '';
    } else if (numAnswer >= question.MaxN + 1) {
      newAnswers.Errors[errorIndex].Message = t('maxSelectError', { max: question.MaxN });
    }

    setAnswers(newAnswers);
  };

  return (
    <div>
      <input
        id={choice}
        className="h-4 w-4 mt-1 mr-2 cursor-pointer accent-indigo-500"
        type="checkbox"
        value={choice}
        checked={isChecked}
        onChange={handleChecks}
      />
      <label htmlFor={choice} className="pl-2 text-gray-600 cursor-pointer">
        {choice}
      </label>
    </div>
  );
};

export { SelectDisplay, SelectHintDisplay };
