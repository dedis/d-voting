import React from 'react';
import { FC } from 'react';
import { RankQuestion } from 'types/configuration';
import ProgressBar from './ProgressBar';

type RankResultProps = {
  rank: RankQuestion;
  rankResult: number[][];
};

// Count and display the results of a rank question.
const RankResult: FC<RankResultProps> = ({ rank, rankResult }) => {
  // Sum the position for each candidate such that a low score is better
  // (e.g if choice 1 ranked first and then fourth then it will have a
  // score of (1-1 + 4-1) = 3) and returns the candidate with the lowest score
  const countBallots = () => {
    const minIndices: number[] = [];
    // the maximum score achievable is (number of choices - 1) * number of ballots
    let min = (rank.Choices.length - 1) * rankResult.length;
    const results = rankResult.reduce((a, b) => {
      return a.map((value, index) => {
        return value + b[index];
      });
    });

    const total = results.reduceRight((a, b) => {
      min = a < min ? a : min;
      min = b < min ? b : min;

      return a + b;
    });

    results.forEach((count, index) => {
      if (count === min) {
        minIndices.push(index);
      }
    });

    return { results, total, minIndices };
  };

  const displayResults = () => {
    const { results, total, minIndices } = countBallots();

    return results.map((res, index) => {
      const percentage = (1 - res / total) * 100;
      const roundedPercentage = (Math.round(percentage * 100) / 100).toString();
      const isBest = minIndices.includes(index);

      return (
        <React.Fragment key={index}>
          <div className="px-4 break-words max-w-xs w-max">
            <span className={`${isBest && 'font-bold'}`}>{rank.Choices[index]}</span>:
          </div>
          <ProgressBar isBest={isBest}>{roundedPercentage}</ProgressBar>
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

export default RankResult;
