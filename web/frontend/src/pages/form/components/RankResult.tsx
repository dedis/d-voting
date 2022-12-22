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
          <div className="px-2 sm:px-4 break-words max-w-xs w-max">
            <span>{rank.Choices[index]}</span>:
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

export const IndividualRankResult: FC<RankResultProps> = ({ rank, rankResult }) => {
  return (
    <div>
      {rankResult[0].map((result, index) => {
        return (
          <React.Fragment key={`rank_${index}`}>
            <div className="flex flex-row px-2 sm:px-4 break-words max-w-xs w-max">
              <div className="mr-2 font-bold">{index + 1}:</div>
              <div>{rank.Choices[rankResult[0].indexOf(index)]}</div>
            </div>
          </React.Fragment>
        );
      })}
    </div>
  );
};

export default RankResult;
