import { FC } from 'react';
import { useTranslation } from 'react-i18next';
import { Answers, SelectQuestion } from 'types/configuration';
import { answersFrom } from 'types/getObjectType';

type SelectProps = {
  select: SelectQuestion;
  answers: Answers;
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>;
};

const Select: FC<SelectProps> = ({ select, answers, setAnswers }) => {
  const { t } = useTranslation();

  const handleChecks = (e: React.ChangeEvent<HTMLInputElement>, choiceIndex: number) => {
    let newAnswers = answersFrom(answers);
    let selectAnswers = newAnswers.SelectAnswers.get(select.ID);

    if (select.MaxN === 1) {
      selectAnswers = new Array<boolean>(select.Choices.length).fill(false);
    }
    // TODO check that this does update the mapping
    selectAnswers[choiceIndex] = e.target.checked;
    newAnswers.SelectAnswers.set(select.ID, selectAnswers);
    let numAnswer = selectAnswers.filter((check: boolean) => check === true).length;

    if (numAnswer >= select.MinN && numAnswer < select.MaxN + 1) {
      newAnswers.Errors.set(select.ID, '');
    } else if (numAnswer >= select.MaxN + 1) {
      newAnswers.Errors.set(select.ID, t('maxSelectError', { max: select.MaxN }));
    }

    setAnswers(newAnswers);
  };

  const hintDisplay = () => {
    let hint = '';
    let max = select.MaxN;
    let min = select.MinN;

    if (max === min) {
      hint =
        max > 1
          ? t('selectMin', { minSelect: min, singularPlural: t('pluralAnswers') })
          : t('selectMin', { minSelect: min, singularPlural: t('singularAnswer') });
    } else if (min === 0) {
      hint =
        max > 1
          ? t('selectMax', { maxSelect: max, singularPlural: t('pluralAnswers') })
          : t('selectMax', { maxSelect: max, singularPlural: t('singularAnswer') });
    } else {
      hint = t('selectBetween', { minSelect: min, maxSelect: max });
    }

    return <div className="text-sm pl-2 pb-2 text-gray-400">{hint}</div>;
  };

  const choiceDisplay = (isChecked: boolean, choice: string, choiceIndex: number) => {
    return (
      <div key={choice}>
        <input
          id={choice}
          className="h-4 w-4 mt-1 mr-2 cursor-pointer accent-indigo-500"
          type="checkbox"
          value={choice}
          checked={isChecked}
          onChange={(e) => handleChecks(e, choiceIndex)}
        />
        <label htmlFor={choice} className="pl-2 text-gray-600 cursor-pointer">
          {choice}
        </label>
      </div>
    );
  };

  return (
    <div>
      <h3 className="text-lg text-gray-600">{select.Title}</h3>
      {hintDisplay()}
      <div className="sm:pl-8 pl-6">
        {Array.from(answers.SelectAnswers.get(select.ID).entries()).map(
          ([choiceIndex, isChecked]) =>
            choiceDisplay(isChecked, select.Choices[choiceIndex], choiceIndex)
        )}
      </div>
      <div className="text-red-600 text-sm py-2 sm:pl-2 pl-1">{answers.Errors.get(select.ID)}</div>
    </div>
  );
};

export default Select;
