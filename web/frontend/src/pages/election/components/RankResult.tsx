import React, { FC } from 'react';
import { RankQuestion } from 'types/configuration';
import ProgressBar from './ProgressBar';
import { countRankResult } from './utils/countResult';

type RankResultProps = {
  rank: RankQuestion;
  rankResult: number[][];
};

// Display the results of a rank question.
const RankResult: FC<RankResultProps> = ({ rank, rankResult }) => {
  const { resultsInPercent, minIndices } = countRankResult(rankResult, rank);

  const displayResults = () => {
    return resultsInPercent.map((percent, index) => {
      const isBest = minIndices.includes(index);

      return (
        <React.Fragment key={index}>
          <div className="px-4 break-words max-w-xs w-max">
            <span className={`${isBest && 'font-bold'}`}>{rank.Choices[index]}</span>:
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

export default RankResult;
