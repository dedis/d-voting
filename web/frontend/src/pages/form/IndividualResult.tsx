import { FC, useEffect, useState } from 'react';
import { RankResults, SelectResults, TextResults } from 'types/form';
import { IndividualSelectResult } from './components/SelectResult';
import RankResult from './components/RankResult';
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
  TextQuestion,
} from 'types/configuration';
import { useParams } from 'react-router-dom';
import useForm from 'components/utils/useForm';
import { useConfigurationOnly } from 'components/utils/useConfiguration';
import Loading from 'pages/Loading';
import { IndividualTextResult } from './components/TextResult';

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
    console.log(element);
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
          <IndividualTextResult
            text={element as TextQuestion}
            textResult={[textResult.get(element.ID)[currentID]]}
          />
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

  const handleBlur = (e) => {
    //console.log(e);
    try {
      const value = parseInt(e.target.value);
    } catch (error) {
      e.target.value = currentID + 1;
      return;
    }
    if (e.target.value > ballotNumber) {
      setCurrentID(ballotNumber - 1);
    } else if (e.target.value < 1) {
      setCurrentID(0);
    } else {
      console.log('in range');
      setCurrentID(e.target.value - 1);
    }
    console.log('ID', currentID);
    e.target.value = currentID + 1;
  };

  const handleEnter = (e) => {
    if (e.key === 'Enter') {
      handleBlur(e);
    }
  };

  // <div className="grow col-span-7 p-2">{'Ballot ' + (currentID + 1)}</div>
  return !loading ? (
    <div>
      <div className="flex flex-col">
        <div className="grid grid-cols-9 font-medium rounded-md border-t stext-sm text-center align-center justify-middle text-gray-700 bg-white py-2">
          <button
            onClick={handlePrevious}
            className="items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
            {t('previous')}
          </button>

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
/*          <input
            type="text"
            min={1}
            max={ballotNumber}
            onBlur={(e) => handleBlur(e)}
            onKeyDown={(e) => handleEnter(e)}
            className="col-span-7 text-center"></input>*/

export default IndividualResult;
