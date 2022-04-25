import React from 'react';
import { FC } from 'react';
import { TextQuestion } from 'types/configuration';
import ProgressBar from './ProgressBar';

type TextResultProps = {
  text: TextQuestion;
  textResult: string[][];
};

// Count and display the results of a text question.
const TextResult: FC<TextResultProps> = ({ text, textResult }) => {
  // Count the number of votes for each candidate and returns the counts and the
  // candidate(s) with the most votes
  const countBallots = () => {
    const results: Map<string, number> = new Map();
    let max = 0;
    const maxKey: string[] = [];

    textResult.forEach((result) => {
      result.forEach((res) => {
        let count = 1;

        if (results.has(res)) {
          count += results.get(res);
        }

        if (count >= max) {
          max = count;
        }
        results.set(res, count);
      });
    });

    results.forEach((count, candidate) => {
      if (count === max) {
        maxKey.push(candidate);
      }
    });

    return { results, maxKey };
  };

  const displayResults = () => {
    const { results, maxKey } = countBallots();

    return Array.from(results.keys()).map((candidate) => {
      const percentage = (results.get(candidate) / textResult.length) * 100;
      const roundedPercentage = (Math.round(percentage * 100) / 100).toFixed(2);
      const isBest = maxKey.includes(candidate);

      return (
        <React.Fragment key={candidate}>
          <div className="px-4 break-words max-w-xs w-max">
            <span className={`${isBest && 'font-bold'}`}>{candidate}</span>:
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

export default TextResult;
