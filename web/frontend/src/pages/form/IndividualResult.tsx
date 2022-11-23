import { FC, useState } from 'react';
import { RankResults, SelectResults, TextResults } from 'types/form';
import SelectResult from './components/SelectResult';
import RankResult from './components/RankResult';
import TextResult from './components/TextResult';
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
};
// Functional component that displays the result of the votes
const IndividualResult: FC<IndividualResultProps> = ({ rankResult, selectResult, textResult }) => {
  const { formId } = useParams();

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

  return (
    <div>
      <div className="flex flex-col">
        {configuration.Scaffold.map((subject: Subject) => displayResults(subject))}
      </div>
    </div>
  );
};

export default IndividualResult;
