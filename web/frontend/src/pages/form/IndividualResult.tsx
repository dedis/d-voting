import { FC, useCallback, useEffect, useState } from 'react';
import {
  BallotResults,
  DownloadedResults,
  RankResults,
  SelectResults,
  TextResults,
} from 'types/form';
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
import { CursorClickIcon, MenuAlt1Icon, SwitchVerticalIcon } from '@heroicons/react/outline';
import DownloadButton from 'components/buttons/DownloadButton';
import { useParams } from 'react-router-dom';
import useForm from 'components/utils/useForm';
import { useConfigurationOnly } from 'components/utils/useConfiguration';
import Loading from 'pages/Loading';
import saveAs from 'file-saver';
import { useNavigate } from 'react-router';
import { default as i18n } from 'i18next';
import { isJson } from 'types/JSONparser';

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
  const { loading, configObj } = useForm(formId);
  const configuration = useConfigurationOnly(configObj);

  const [currentID, setCurrentID] = useState<string>('1');
  const [isValid, setIsValid] = useState<ValidityType>(0);
  const [internalID, setInternalID] = useState<number>(0);

  enum ValidityType {
    VALID = 0,
    UNPARSABLE = 1,
    OUT_OF_BOUNDS = 2,
  }

  const SubjectElementResultDisplay = useCallback(
    (element: SubjectElement) => {
      const questionIcons = {
        [RANK]: <SwitchVerticalIcon />,
        [SELECT]: <CursorClickIcon />,
        [TEXT]: <MenuAlt1Icon />,
      };

      let titles;
      if (isJson(element.Title)) {
        titles = JSON.parse(element.Title);
      }
      if (titles === undefined) {
        titles = { en: element.Title, fr: element.TitleFr, de: element.TitleDe };
      }
      return (
        <div className="pl-4 pb-4 sm:pl-6 sm:pb-6">
          <div className="flex flex-row">
            <div className="align-text-middle flex mt-1 mr-2 h-5 w-5" aria-hidden="true">
              {questionIcons[element.Type]}
            </div>
            <h2 className="flex align-text-middle text-lg pb-2">
              {i18n.language === 'en' && titles.en}
              {i18n.language === 'fr' && titles.fr}
              {i18n.language === 'de' && titles.de}
            </h2>
          </div>
          {element.Type === RANK && rankResult.has(element.ID) && (
            <IndividualRankResult
              rank={element as RankQuestion}
              rankResult={[rankResult.get(element.ID)[internalID]]}
            />
          )}
          {element.Type === SELECT && selectResult.has(element.ID) && (
            <IndividualSelectResult
              select={element as SelectQuestion}
              selectResult={[selectResult.get(element.ID)[internalID]]}
            />
          )}
          {element.Type === TEXT && textResult.has(element.ID) && (
            <IndividualTextResult
              text={element as TextQuestion}
              textResult={[textResult.get(element.ID)[internalID]]}
            />
          )}
        </div>
      );
    },
    [internalID, rankResult, selectResult, textResult]
  );

  const displayResults = useCallback(
    (subject: Subject) => {
      let sbj;
      if (isJson(subject.Title)) {
        sbj = JSON.parse(subject.Title);
      }
      if (sbj === undefined) {
        sbj = { en: subject.Title, fr: subject.TitleFr, de: subject.TitleDe };
      }
      return (
        <div key={subject.ID}>
          <h2 className="text-xl pt-1 pb-1 sm:pt-2 sm:pb-2 border-t font-bold text-gray-600">
            {i18n.language === 'en' && sbj.en}
            {i18n.language === 'fr' && sbj.fr}
            {i18n.language === 'de' && sbj.de}
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
    },
    [SubjectElementResultDisplay]
  );

  const getResultData = (
    subject: Subject,
    dataToDownload: DownloadedResults[],
    BallotID: number
  ) => {
    dataToDownload.push({ Title: subject.Title });

    subject.Order.forEach((id: ID) => {
      const element = subject.Elements.get(id);
      let res = undefined;

      switch (element.Type) {
        case RANK:
          const rankQues = element as RankQuestion;

          if (rankResult.has(id)) {
            res = rankResult.get(id)[BallotID].map((rank, index) => {
              return {
                Rank: `${index + 1}`,
                Choice: rankQues.Choices[rankResult.get(id)[BallotID].indexOf(index)],
              };
            });
            dataToDownload.push({ Title: element.Title, Results: res });
          }
          break;

        case SELECT:
          const selectQues = element as SelectQuestion;

          if (selectResult.has(id)) {
            res = selectResult.get(id)[BallotID].map((select, index) => {
              const checked = select ? 'True' : 'False';
              return { Candidate: selectQues.Choices[index], Checked: checked };
            });
            dataToDownload.push({ Title: element.Title, Results: res });
          }
          break;

        case SUBJECT:
          getResultData(element as Subject, dataToDownload, BallotID);
          break;

        case TEXT:
          const textQues = element as TextQuestion;

          if (textResult.has(id)) {
            res = textResult.get(id)[BallotID].map((text, index) => {
              return { Field: textQues.Choices[index], Answer: text };
            });
            dataToDownload.push({ Title: element.Title, Results: res });
          }
          break;
      }
    });
  };

  const exportJSONData = () => {
    const fileName = `result_${configuration.MainTitle.replace(/[^a-zA-Z0-9]/g, '_').slice(
      0,
      99
    )}__individual`;
    const ballotsToDownload: BallotResults[] = [];

    const indices: number[] = [...Array(ballotNumber).keys()];
    indices.forEach((BallotID) => {
      const dataToDownload: DownloadedResults[] = [];
      configuration.Scaffold.forEach((subject: Subject) => {
        getResultData(subject, dataToDownload, BallotID);
      });
      ballotsToDownload.push({ BallotNumber: BallotID + 1, Results: dataToDownload });
    });

    const data = {
      MainTitle: configuration.MainTitle,
      TitleFr: configuration.TitleFr,
      TitleDe: configuration.TitleDe,
      NumberOfVotes: ballotNumber,
      Ballots: ballotsToDownload,
    };

    const fileToSave = new Blob([JSON.stringify(data, null, 2)], {
      type: 'application/json',
    });

    saveAs(fileToSave, fileName);
  };

  useEffect(() => {
    let value: number;
    value = parseInt(currentID);
    if (isNaN(value)) {
      setIsValid(ValidityType.UNPARSABLE);
    } else if (value < 1 || value > ballotNumber) {
      setIsValid(ValidityType.OUT_OF_BOUNDS);
    } else {
      setIsValid(ValidityType.VALID);
      setInternalID(value - 1);
    }
  }, [ValidityType, ballotNumber, currentID]);

  useEffect(() => {
    configuration.Scaffold.map((subject: Subject) => displayResults(subject));
  }, [configuration, displayResults]);

  const handleNext = (): void => {
    setCurrentID((((internalID + 1) % ballotNumber) + 1).toString());
  };

  const handlePrevious = (): void => {
    setCurrentID((((internalID - 1 + ballotNumber) % ballotNumber) + 1).toString());
  };

  return !loading ? (
    <div>
      <div className="flex flex-col">
        <div className="grid grid-cols-9 font-medium rounded-md border-t stext-sm text-center align-center justify-middle text-gray-700 bg-white py-2">
          <button
            onClick={handlePrevious}
            className="col-span-2 items-center mx-3 px-2 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
            {t('previous')}
          </button>
          <div className="flex flex-col align-middle col-span-5 col-start-3">
            <input
              type="text"
              inputMode="numeric"
              pattern="[0-9]*"
              title="Please enter a number in the range of ballot numbers"
              onChange={(e) => {
                setCurrentID(e.target.value);
              }}
              className={
                'grow align-middle text-center border border-gray-300 rounded-md ring-transparent focus:ring-0' +
                (isValid !== 0 ? ' border-red-500 border-1 text-red-600' : '')
              }
              value={currentID}
            />
            {isValid !== 0 && (
              <p role="alert" className="py-0 text-left text-xs text-red-600">
                {t('invalidInput', { max: ballotNumber })}
              </p>
            )}
          </div>

          <button
            onClick={handleNext}
            className="col-span-2 mx-3 relative align-right items-center px-2 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
            {t('next')}
          </button>
        </div>
        {configuration.Scaffold.map((subject: Subject) => displayResults(subject))}
      </div>
      <div className="flex my-4">
        <button
          type="button"
          onClick={() => navigate(-1)}
          className="text-gray-700 my-2 mr-2 items-center px-4 py-2 border rounded-md text-sm hover:text-[#ff0000]">
          {t('back')}
        </button>

        <DownloadButton exportData={exportJSONData}>{t('exportJSON')}</DownloadButton>
      </div>
    </div>
  ) : (
    <Loading />
  );
};

export default IndividualResult;
