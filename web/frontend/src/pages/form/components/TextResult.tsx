import React, { FC } from 'react';
import { TextQuestion } from 'types/configuration';
import ProgressBar from './ProgressBar';
import { countTextResult } from './utils/countResult';
import { default as i18n } from 'i18next';

type TextResultProps = {
  textResult: string[][];
};
type IndividualTextResultProps = {
  text: TextQuestion;
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

export const IndividualTextResult: FC<IndividualTextResultProps> = ({ text, textResult }) => {
  return (
    <div>
      {textResult[0].map((result, index) => {
        return (
          <React.Fragment key={`txt_${index}`}>
            <div className="flex flex-row px-2 sm:px-4 break-words max-w-xs w-max">
              <div className="mr-2 font-bold"> 
              {i18n.language === 'en' && text.ChoicesMap.get('en')[index]}
              {i18n.language === 'fr' && text.ChoicesMap.get('fr')[index]}
              {i18n.language === 'de' && text.ChoicesMap.get('de')[index]}:
              </div>
              <div>{result}</div>
            </div>
          </React.Fragment>
        );
      })}
    </div>
  );
};

export default TextResult;
