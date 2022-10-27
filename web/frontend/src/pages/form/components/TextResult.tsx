import React, { FC } from 'react';
import ProgressBar from './ProgressBar';
import { countTextResult } from './utils/countResult';

type TextResultProps = {
  textResult: string[][];
};

// Display the results of a text question.
const TextResult: FC<TextResultProps> = ({ textResult }) => {
  const { resultsInPercent, maxKey } = countTextResult(textResult);

  const displayResults = () => {
    return Array.from(resultsInPercent).map(([textAnswer, result]) => {
      const isBest = maxKey.includes(textAnswer);

      return (
        <React.Fragment key={textAnswer}>
          <div className="px-2 sm:px-4 break-words max-w-xs w-max">
            <span>{textAnswer}</span>:
          </div>
          <ProgressBar isBest={isBest}>{result}</ProgressBar>
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
