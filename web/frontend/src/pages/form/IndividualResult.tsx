import { FC, useEffect, useState } from 'react';
import { DownloadedResults, RankResults, SelectResults, TextResults } from 'types/form';
import { IndividualSelectResult } from './components/SelectResult';
import { IndividualTextResult } from './components/TextResult';
import { IndividualRankResult } from './components/RankResult';
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
import DownloadButton from 'components/buttons/DownloadButton';
import { useParams } from 'react-router-dom';
import useForm from 'components/utils/useForm';
import { useConfigurationOnly } from 'components/utils/useConfiguration';
import Loading from 'pages/Loading';
import saveAs from 'file-saver';
import { useNavigate } from 'react-router';

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
  const navigate = useNavigate();
  const { loading, result, configObj } = useForm(formId);
  const configuration = useConfigurationOnly(configObj);

  const [currentID, setCurrentID] = useState<number>(0);

  const SubjectElementResultDisplay = (element: SubjectElement) => {
    return (
      <div className="pl-4 pb-4 sm:pl-6 sm:pb-6">
        <h2 className="text-lg pb-2">{element.Title}</h2>
        {element.Type === RANK && rankResult.has(element.ID) && (
          <IndividualRankResult
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

  const getResultData = (subject: Subject, dataToDownload: DownloadedResults[]) => {
    dataToDownload.push({ Title: subject.Title });

    subject.Order.forEach((id: ID) => {
      const element = subject.Elements.get(id);
      let res = undefined;

      switch (element.Type) {
        case RANK:
          const rankQues = element as RankQuestion;

          if (rankResult.has(id)) {
            res = rankResult.get(id)[currentID].map((rank, index) => {
              return {
                Placement: `${index + 1}`,
                Holder: rankQues.Choices[rankResult.get(id)[currentID].indexOf(index)],
              };
            });
            dataToDownload.push({ Title: element.Title, Results: res });
          }
          break;

        case SELECT:
          const selectQues = element as SelectQuestion;

          if (selectResult.has(id)) {
            res = selectResult.get(id)[currentID].map((select, index) => {
              const checked = select ? 'True' : 'False';
              return { Candidate: selectQues.Choices[index], Checked: checked };
            });
            dataToDownload.push({ Title: element.Title, Results: res });
          }
          break;

        case SUBJECT:
          getResultData(element as Subject, dataToDownload);
          break;

        case TEXT:
          const textQues = element as TextQuestion;

          if (textResult.has(id)) {
            res = textResult.get(id)[currentID].map((text, index) => {
              return { Field: textQues.Choices[index], Answer: text };
            });
            dataToDownload.push({ Title: element.Title, Results: res });
          }
          break;
      }
    });
  };

  const exportJSONData = () => {
    const fileName = `result_${configuration.MainTitle}_Ballot${currentID + 1}.json`;

    const dataToDownload: DownloadedResults[] = [];

    configuration.Scaffold.forEach((subject: Subject) => {
      getResultData(subject, dataToDownload);
    });

    const data = {
      Title: configuration.MainTitle,
      BallotNumber: currentID + 1,
      Results: dataToDownload,
    };

    const fileToSave = new Blob([JSON.stringify(data, null, 2)], {
      type: 'application/json',
    });

    saveAs(fileToSave, fileName);
  };

  useEffect(() => {
    configuration.Scaffold.map((subject: Subject) => displayResults(subject));
  }, [currentID]);

  const handleNext = (): void => {
    setCurrentID(currentID + 1);
  };

  const handlePrevious = (): void => {
    setCurrentID(currentID - 1);
  };

  const handleBlur = (e) => {
    setCurrentID(e.target.value - 1);
  };

  const handleEnter = (e) => {
    switch (e.key) {
      case 'Enter':
        handleBlur(e);
        break;
      case 'ArrowUp':
        handleNext();
        break;
      case 'ArrowDown':
        handlePrevious();
        break;
    }
  };

  // <div className="grow col-span-7 p-2">{'Ballot ' + (currentID + 1)}</div>
  return !loading ? (
    <div>
      <div className="flex flex-col">
        <div className="grid grid-cols-9 font-medium rounded-md border-t stext-sm text-center align-center justify-middle text-gray-700 bg-white py-2">
          <input
            type="text"
            inputMode="numeric"
            pattern="[0-9]*"
            title="Please enter a number"
            onChange={(e) => console.log(e.target.value)}
            onBlur={(e) => handleBlur(e)}
            onKeyDown={(e) => handleEnter(e)}
            className="col-span-7 col-start-2 text-center"
            value={1}
          />
        </div>
        {configuration.Scaffold.map((subject: Subject) => displayResults(subject))}
      </div>
      <div className="flex my-4">
        <button
          type="button"
          onClick={() => navigate(-1)}
          className="text-gray-700 my-2 mr-2 items-center px-4 py-2 border rounded-md text-sm hover:text-indigo-500">
          {t('back')}
        </button>

        <DownloadButton exportData={exportJSONData}>{t('exportJSON')}</DownloadButton>
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
