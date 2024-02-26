import React, { FC } from 'react';
import { RankQuestion } from 'types/configuration';
import { ProgressBar } from './ProgressBar';
import { countRankResult } from './utils/countResult';
import { default as i18n } from 'i18next';

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
            <span>
              {i18n.language === 'en' && rank.ChoicesMap.ChoicesMap.get('en')[index]}
              {i18n.language === 'fr' && rank.ChoicesMap.ChoicesMap.get('fr')[index]}
              {i18n.language === 'de' && rank.ChoicesMap.ChoicesMap.get('de')[index]}
            </span>
            :
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
              <div>
                {i18n.language === 'en' &&
                  rank.ChoicesMap.ChoicesMap.get('en')[rankResult[0].indexOf(index)]}
                {i18n.language === 'fr' &&
                  rank.ChoicesMap.ChoicesMap.get('fr')[rankResult[0].indexOf(index)]}
                {i18n.language === 'de' &&
                  rank.ChoicesMap.ChoicesMap.get('de')[rankResult[0].indexOf(index)]}
              </div>
            </div>
          </React.Fragment>
        );
      })}
    </div>
  );
};

export default RankResult;
