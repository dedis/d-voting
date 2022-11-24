import { FC, useEffect, useState } from 'react';
import { RankResults, SelectResults, TextResults } from 'types/form';
import SelectResult from './components/SelectResult';
import RankResult from './components/RankResult';
import TextResult from './components/TextResult';
import { useTranslation } from 'react-i18next';
import {
  ID,
  RANK,
  RankQuestion,
  SELECT,
  SUBJECT,
  SelectQuestion,
  Subject,
  SubjectElement,
  TEXT,
} from 'types/configuration';
import { useParams } from 'react-router-dom';
import useForm from 'components/utils/useForm';
import { useConfigurationOnly } from 'components/utils/useConfiguration';

type IndividualResultProps = {
  rankResult: RankResults;
  selectResult: SelectResults;
  textResult: TextResults;
  ballotNumber: number;
};
// Functional component that displays the result of the votes
const IndividualResult: FC<IndividualResultProps> = ({
  rankResult,
  selectResult,
  textResult,
  ballotNumber,
}) => {
  const { formId } = useParams();
  const { t } = useTranslation();
  const { configObj } = useForm(formId);
  const configuration = useConfigurationOnly(configObj);

  const [currentID, setCurrentID] = useState<number>(0);

  const SubjectElementResultDisplay = (element: SubjectElement) => {
    return (
      <div className="pl-4 pb-4 sm:pl-6 sm:pb-6">
        <h2 className="text-lg pb-2">{element.Title}</h2>
        {element.Type === RANK && rankResult.has(element.ID) && (
          <RankResult
            rank={element as RankQuestion}
            rankResult={[rankResult.get(element.ID)[currentID]]}
          />
        )}
        {element.Type === SELECT && selectResult.has(element.ID) && (
          <SelectResult
            select={element as SelectQuestion}
            selectResult={[selectResult.get(element.ID)[currentID]]}
          />
        )}
        {element.Type === TEXT && textResult.has(element.ID) && (
          <TextResult textResult={[textResult.get(element.ID)[currentID]]} />
        )}
      </div>
    );
  };

  const displayResults = (subject: Subject) => {
    return (
      <div key={subject.ID}>
        <h2 className="text-xl pt-1 pb-1 sm:pt-2 sm:pb-2 border-t font-bold text-gray-600">
          {subject.Title}
        </h2>
        {subject.Order.map((id: ID) => (
          <div key={id}>
            {subject.Elements.get(id).Type === SUBJECT ? (
              <div className="pl-4 sm:pl-6">
                {displayResults(subject.Elements.get(id) as Subject)}
              </div>
            ) : (
              SubjectElementResultDisplay(subject.Elements.get(id))
            )}
          </div>
        ))}
      </div>
    );
  };
  useEffect(() => {
    configuration.Scaffold.map((subject: Subject) => displayResults(subject));
  }, [currentID]);

  const handleNext = (): void => {
    setCurrentID((currentID + 1) % ballotNumber);
  };

  const handlePrevious = (): void => {
    setCurrentID((currentID - 1) % ballotNumber);
  };
  return (
    <div>
      <div className="flex flex-col">
        <div className="flex-1 flex justify-between sm:justify-end">
          <button
            onClick={handlePrevious}
            className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
            {t('previous')}
          </button>
          {'Ballot ' + (currentID + 1)}
          <button
            onClick={handleNext}
            className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
            {t('next')}
          </button>
        </div>
        {configuration.Scaffold.map((subject: Subject) => displayResults(subject))}
      </div>
    </div>
  );
};

export default IndividualResult;
