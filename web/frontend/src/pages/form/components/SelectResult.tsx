import React, { FC } from 'react';
import { SelectQuestion } from 'types/configuration';
import ProgressBar from './ProgressBar';
import { countSelectResult } from './utils/countResult';
import { default as i18n } from 'i18next';

type SelectResultProps = {
  select: SelectQuestion;
  selectResult: number[][];
};

// Display the results of a select question.
const SelectResult: FC<SelectResultProps> = ({ select, selectResult }) => {
  const { resultsInPercent, maxIndices } = countSelectResult(selectResult);

  const displayResults = () => {
    return resultsInPercent.map((percent, index) => {
      const isBest = maxIndices.includes(index);

      return (
        <React.Fragment key={index}>
          <div className="px-2 sm:px-4 break-words max-w-xs w-max">
            <span>
              {i18n.language === 'en' && select.ChoicesMap.get('en')[index]}
              {i18n.language === 'fr' && select.ChoicesMap.get('fr')[index]}
              {i18n.language === 'de' && select.ChoicesMap.get('de')[index]}
            </span>
            :
          </div>
          <ProgressBar isBest={isBest}>{percent}</ProgressBar>
        </React.Fragment>
      );
    });
  };

  return (
    <div className="grid [grid-template-columns:_min-content_auto] gap-1 items-center">
      {displayResults()}
    </div>
  );
};

export const IndividualSelectResult: FC<SelectResultProps> = ({ select, selectResult }) => {
  const displayChoices = (result, index) => {
    return (
      <div>
        <input type="checkbox" key={index} checked={result} disabled />
      </div>
    );
  };
  return (
    <div>
      {selectResult[0].map((result, index) => {
        return (
          <React.Fragment key={`select_${index}`}>
            <div className="flex flex-row px-2 sm:px-4 break-words max-w-xs w-max">
              <div className="h-4 w-4 mr-2 accent-indigo-500 ">{displayChoices(result, index)}</div>
              <div>
                {i18n.language === 'en' && select.ChoicesMap.get('en')[index]}
                {i18n.language === 'fr' && select.ChoicesMap.get('fr')[index]}
                {i18n.language === 'de' && select.ChoicesMap.get('de')[index]}
              </div>
            </div>
          </React.Fragment>
        );
      })}
    </div>
  );
};

export default SelectResult;
