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
    return Array.from(resultsInPercent).map((res) => {
      const isBest = maxKey.includes(res[0]);

      return (
        <React.Fragment key={res[0]}>
          <div className="px-4 break-words max-w-xs w-max">
            <span className={`${isBest && 'font-bold'}`}>{res[0]}</span>:
          </div>
          <ProgressBar isBest={isBest}>{res[1]}</ProgressBar>
        </React.Fragment>
      );
    });
  };

  return (
    <div className="grid [grid-template-columns:_min-content_auto] gap-1 items-center w-4/5">
      {displayResults()}
    </div>
  );
};

export default TextResult;
