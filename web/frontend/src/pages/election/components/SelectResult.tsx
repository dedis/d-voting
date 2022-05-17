import React, { FC } from 'react';
import { SelectQuestion } from 'types/configuration';
import ProgressBar from './ProgressBar';
import { countSelectResult } from './utils/countResult';

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
          <div className="px-4 break-words max-w-xs w-max">
            <span className={`${isBest && 'font-bold'}`}>{select.Choices[index]}</span>:
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

export default SelectResult;
