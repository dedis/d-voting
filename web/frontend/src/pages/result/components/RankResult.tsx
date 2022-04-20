import { FC } from 'react';
import { RankQuestion } from 'types/configuration';
import ProgressBar from './ProgressBar';

type RankResultProps = {
  rank: RankQuestion;
  rankResult: number[][];
};

// Count and display the results of a rank question.
const RankResult: FC<RankResultProps> = ({ rank, rankResult }) => {
  const countBallots = () => {
    let minIndices: number[] = [];
    // max is number of choices * number of ballots
    let min = rankResult.length * rank.Choices.length;
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

    console.log('total: ' + total);
    console.log('min: ' + min);
    console.log('minIndices: ' + minIndices);
    console.log('results: ' + results);
    return { results, total, minIndices };
  };

  const displayResults = () => {
    const { results, total, minIndices } = countBallots();

    return results.map((res, index) => {
      const percentage = (1 - res / total) * 100;
      const roundedPercentage = (Math.round(percentage * 100) / 100).toString();

      return (
        <div key={index}>
          <div>
            <ProgressBar candidate={rank.Choices[index]} isBest={minIndices.includes(index)}>
              {roundedPercentage}
            </ProgressBar>
          </div>
        </div>
      );
    });
  };

  return <div>{displayResults()}</div>;
};

export default RankResult;
