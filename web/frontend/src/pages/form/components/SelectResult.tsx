import React, { FC } from 'react';
import { SelectQuestion } from 'types/configuration';
import { SelectProgressBar } from './ProgressBar';
import { countSelectResult } from './utils/countResult';
import { prettifyChoice } from './utils/display';

type SelectResultProps = {
  select: SelectQuestion;
  selectResult: number[][];
};

// Display the results of a select question.
const SelectResult: FC<SelectResultProps> = ({ select, selectResult }) => {
  const sortedResults = countSelectResult(selectResult)
    .map((result, index) => {
      const tempResult: [string, number, number] = [...result, index];
      return tempResult;
    })
    .sort((x, y) => y[1] - x[1]);
  const maxCount = sortedResults[0][1];

  const displayResults = () => {
    return sortedResults.map(([percent, totalCount, origIndex], index) => {
      return (
        <React.Fragment key={index}>
          <div className="px-2 sm:px-4 break-words max-w-xs w-max">
            <span>{prettifyChoice(select.ChoicesMap, origIndex)}</span>:
          </div>
          <SelectProgressBar
            percent={percent}
            totalCount={totalCount}
            numberOfBallots={selectResult.length}
            isBest={totalCount === maxCount}></SelectProgressBar>
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
              <div className="h-4 w-4 mr-2 accent-[#ff0000] ">{displayChoices(result, index)}</div>
              <div>{prettifyChoice(select.ChoicesMap, index)}</div>
            </div>
          </React.Fragment>
        );
      })}
    </div>
  );
};

export default SelectResult;
