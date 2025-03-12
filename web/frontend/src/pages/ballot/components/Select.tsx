import React, { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Answers, SelectQuestion } from 'types/configuration';
import { answersFrom } from 'types/getObjectType';
import HintButton from 'components/buttons/HintButton';
import { internationalize, urlizeLabel } from './../../utils';
type SelectProps = {
  select: SelectQuestion;
  answers: Answers;
  setAnswers: (answers: Answers) => void;
  language: string;
};

const Select: FC<SelectProps> = ({ select, answers, setAnswers, language }) => {
  const { t } = useTranslation();
  const handleChecks = (e: React.ChangeEvent<HTMLInputElement>, choiceIndex: number) => {
    const newAnswers = answersFrom(answers);
    let selectAnswers = newAnswers.SelectAnswers.get(select.ID);

    if (select.MaxN === 1) {
      selectAnswers = new Array<boolean>(select.Choices.length).fill(false);
    }

    selectAnswers[choiceIndex] = e.target.checked;
    newAnswers.SelectAnswers.set(select.ID, selectAnswers);
    const numAnswer = selectAnswers.filter((check: boolean) => check === true).length;

    if (numAnswer >= select.MinN && numAnswer < select.MaxN + 1) {
      newAnswers.Errors.set(select.ID, '');
    } else if (numAnswer >= select.MaxN + 1) {
      newAnswers.Errors.set(select.ID, t('maxSelectError', { max: select.MaxN }));
    }

    setAnswers(newAnswers);
  };

  const requirementsDisplay = () => {
    let requirements;
    const max = select.MaxN;
    const min = select.MinN;

    if (max === min) {
      requirements =
        max > 1
          ? t('selectMin', { minSelect: min, singularPlural: t('pluralAnswers') })
          : t('selectMin', { minSelect: min, singularPlural: t('singularAnswer') });
    } else if (min === 0) {
      requirements =
        max > 1
          ? t('selectMax', { maxSelect: max, singularPlural: t('pluralAnswers') })
          : t('selectMax', { maxSelect: max, singularPlural: t('singularAnswer') });
    } else {
      requirements = t('selectBetween', { minSelect: min, maxSelect: max });
    }

    return <div className="text-sm pl-2 pb-2 sm:pl-4 text-gray-400">{requirements}</div>;
  };
  const [titles, setTitles] = useState<any>({});
  useEffect(() => {
    setTitles(select.Title);
  }, [select]);

  const choiceDisplay = (isChecked: boolean, choice: string, url: string, choiceIndex: number) => {
    const prettyChoice = urlizeLabel(choice, url);
    return (
      <div key={choice}>
        <input
          id={choice}
          className="h-4 w-4 mt-1 mr-2 cursor-pointer accent-[#ff0000]"
          type="checkbox"
          value={choice}
          checked={isChecked}
          onChange={(e) => handleChecks(e, choiceIndex)}
        />
        <label htmlFor={choice} className="pl-2 break-words text-gray-600 cursor-pointer">
          {prettyChoice}
        </label>
      </div>
    );
  };
  return (
    <div>
      <div className="grid grid-rows-1 grid-flow-col">
        <div>
          <h3 className="text-lg break-words text-gray-600">
            {urlizeLabel(internationalize(language, titles), titles.URL)}
          </h3>
        </div>
        <div className="text-right">
          {<HintButton text={internationalize(language, select.Hint)} />}
        </div>
      </div>
      <div className="pt-1">{requirementsDisplay()}</div>
      <div className="sm:pl-8 mt-2 pl-6">
        {Array.from(answers.SelectAnswers.get(select.ID).entries())
          .map(([choiceIndex, isChecked]) => {
            if (language === 'fr' && select.ChoicesMap.ChoicesMap.has('fr'))
              return choiceDisplay(
                isChecked,
                select.ChoicesMap.ChoicesMap.get('fr')[choiceIndex],
                select.ChoicesMap.URLs[choiceIndex],
                choiceIndex
              );
            else if (language === 'de' && select.ChoicesMap.ChoicesMap.has('de'))
              return choiceDisplay(
                isChecked,
                select.ChoicesMap.ChoicesMap.get('de')[choiceIndex],
                select.ChoicesMap.URLs[choiceIndex],
                choiceIndex
              );
            else if (select.ChoicesMap.ChoicesMap.has('en'))
              return choiceDisplay(
                isChecked,
                select.ChoicesMap.ChoicesMap.get('en')[choiceIndex],
                select.ChoicesMap.URLs[choiceIndex],
                choiceIndex
              );
            return undefined;
          })
          .filter((e) => e !== undefined)}
      </div>
      <div className="text-red-600 text-sm py-2 sm:pl-4 pl-2">{answers.Errors.get(select.ID)}</div>
    </div>
  );
};

export default Select;
