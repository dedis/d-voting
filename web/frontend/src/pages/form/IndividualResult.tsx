import { FC, useEffect, useState } from 'react';
import { RankResults, SelectResults, TextResults } from 'types/form';
import { IndividualSelectResult } from './components/SelectResult';
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
import Loading from 'pages/Loading';

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
  const { loading, configObj } = useForm(formId);
  const configuration = useConfigurationOnly(configObj);

  const [currentID, setCurrentID] = useState<number>(0);

  const SubjectElementResultDisplay = (element: SubjectElement) => {
    console.log(element.Type == SELECT ? selectResult.get(element.ID) : 'none');
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
          <IndividualSelectResult
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
    setCurrentID((currentID - 1 + ballotNumber) % ballotNumber);
  };

  return !loading ? (
    <div>
      <div className="flex flex-col">
        <div className="grid grid-cols-9 font-medium rounded-md border-t stext-sm text-center align-center justify-middle text-gray-700 bg-white py-2">
          <button
            onClick={handlePrevious}
            className="items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
            {t('previous')}
          </button>
          <div className="grow col-span-7 p-2">{'Ballot ' + (currentID + 1)}</div>
          <button
            onClick={handleNext}
            className="ml-3 relative align-right items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
            {t('next')}
          </button>
        </div>
        {configuration.Scaffold.map((subject: Subject) => displayResults(subject))}
      </div>
    </div>
  ) : (
    <Loading />
  );
};

export default IndividualResult;
