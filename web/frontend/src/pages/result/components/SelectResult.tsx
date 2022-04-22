import React from 'react';
import { FC } from 'react';
import { SelectQuestion } from 'types/configuration';
import ProgressBar from './ProgressBar';

type SelectResultProps = {
  select: SelectQuestion;
  selectResult: number[][];
};

// Count and display the results of a select question.
const SelectResult: FC<SelectResultProps> = ({ select, selectResult }) => {
  // Count the number of vote for a candidate and returns which candidates
  // in the select.Choices has the most votes
  const countBallots = () => {
    const maxIndices: number[] = [];
    let max = 0;
    const results = selectResult.reduce((a, b) => {
      return a.map((value, index) => {
        const current = value + b[index];

        if (current >= max) {
          max = current;
        }
        return current;
      });
    });

    results.forEach((count, index) => {
      if (count === max) {
        maxIndices.push(index);
      }
    });
    return { results, maxIndices };
  };

  const { results, maxIndices } = countBallots();

  const displayResults = () => {
    return results.map((res, index) => {
      const percentage = (res / selectResult.length) * 100;
      const roundedPercentage = (Math.round(percentage * 100) / 100).toFixed(2);
      const isBest = maxIndices.includes(index);

      return (
        <React.Fragment key={index}>
          <div className="px-4 break-words max-w-xs w-max">
            <span className={`${isBest && 'font-bold'}`}>{select.Choices[index]}</span>:
          </div>
          <ProgressBar isBest={isBest}>{roundedPercentage}</ProgressBar>
        </React.Fragment>
      );
    });
  };

  return (
    <div
      className="grid [grid-template-columns:_min-content_auto] gap-1 items-center"
      key={select.ID}>
      {displayResults()}
    </div>
  );
};

export default SelectResult;
