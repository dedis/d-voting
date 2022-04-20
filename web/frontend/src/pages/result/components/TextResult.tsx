import { FC } from 'react';
import { TextQuestion } from 'types/configuration';
import ProgressBar from './ProgressBar';

type TextResultProps = {
  text: TextQuestion;
  textResult: string[][];
};

// Count and display the results of a text question.
const TextResult: FC<TextResultProps> = ({ text, textResult }) => {
  const countBallots = () => {
    const results: Map<string, number> = new Map();
    let max = 0;
    let maxKey: string[] = [];

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

      return (
        <div key={candidate}>
          <div>
            <ProgressBar candidate={candidate} isBest={maxKey.includes(candidate)}>
              {roundedPercentage}
            </ProgressBar>
          </div>
        </div>
      );
    });
  };

  return <div>{displayResults()}</div>;
};

export default TextResult;
