import { FC } from 'react';
import { SelectQuestion } from 'types/configuration';
import ProgressBar from './ProgressBar';

type SelectResultProps = {
  select: SelectQuestion;
  selectResult: number[][];
};

// Count and display the results of a select question.
const SelectResult: FC<SelectResultProps> = ({ select, selectResult }) => {
  const countBallots = () => {
    let maxIndices: number[] = [];
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
    console.log('results: ' + results);

    results.forEach((count, index) => {
      if (count === max) {
        maxIndices.push(index);
      }
    });
    return { results, maxIndices };
  };

  const displayResults = () => {
    const { results, maxIndices } = countBallots();
    //const maxIndex = results.indexOf(Math.max(...results));

    return results.map((res, index) => {
      const percentage = (res / selectResult.length) * 100;
      const roundedPercentage = (Math.round(percentage * 100) / 100).toFixed(2);

      return (
        <div key={index}>
          <div>
            <ProgressBar candidate={select.Choices[index]} isBest={maxIndices.includes(index)}>
              {roundedPercentage}
            </ProgressBar>
          </div>
        </div>
      );
    });
  };

  return <div>{displayResults()}</div>;
};

export default SelectResult;
