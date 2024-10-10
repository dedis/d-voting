import { FC } from 'react';
import { useTranslation } from 'react-i18next';
import { DownloadedResults, RankResults, SelectResults, TextResults } from 'types/form';
import RankResult from './components/RankResult';
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
import { useNavigate } from 'react-router';
import { useConfigurationOnly } from 'components/utils/useConfiguration';
import DownloadButton from 'components/buttons/DownloadButton';
import Loading from 'pages/Loading';
import saveAs from 'file-saver';
import { CursorClickIcon, MenuAlt1Icon, SwitchVerticalIcon } from '@heroicons/react/outline';
import {
  countRankResult,
  countSelectResult,
  countTextResult,
} from './components/utils/countResult';
import { default as i18n } from 'i18next';
import SelectResult from './components/SelectResult';
import TextResult from './components/TextResult';
import { internationalize, urlizeLabel } from './../utils';

type GroupedResultProps = {
  rankResult: RankResults;
  selectResult: SelectResults;
  textResult: TextResults;
};

// Functional component that displays the result of the votes
const GroupedResult: FC<GroupedResultProps> = ({ rankResult, selectResult, textResult }) => {
  const { formId } = useParams();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { loading, result, configObj } = useForm(formId);
  const configuration = useConfigurationOnly(configObj);

  const questionIcons = {
    [RANK]: <SwitchVerticalIcon />,
    [SELECT]: <CursorClickIcon />,
    [TEXT]: <MenuAlt1Icon />,
  };

  const SubjectElementResultDisplay = (element: SubjectElement) => {
    let titles;
    titles = element.Title;
    return (
      <div className="pl-4 pb-4 sm:pl-6 sm:pb-6">
        <div className="flex flex-row">
          <div className="align-text-middle flex mt-1 mr-2 h-5 w-5" aria-hidden="true">
            {questionIcons[element.Type]}
          </div>
          <h2 className="text-lg pb-2">{internationalize(i18n.language, titles)}</h2>
        </div>
        {element.Type === RANK && rankResult.has(element.ID) && (
          <RankResult rank={element as RankQuestion} rankResult={rankResult.get(element.ID)} />
        )}
        {element.Type === SELECT && selectResult.has(element.ID) && (
          <SelectResult
            select={element as SelectQuestion}
            selectResult={selectResult.get(element.ID)}
          />
        )}
        {element.Type === TEXT && textResult.has(element.ID) && (
          <TextResult textResult={textResult.get(element.ID)} />
        )}
      </div>
    );
  };

  const displayResults = (subject: Subject) => {
    return (
      <div key={subject.ID}>
        <h2 className="text-xl pt-1 pb-1 sm:pt-2 sm:pb-2 border-t font-bold text-gray-600">
          {urlizeLabel(internationalize(i18n.language, subject.Title), subject.Title.URL)}
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
    dataToDownload.push({ Title: subject.Title.En, URL: subject.Title.URL });

    subject.Order.forEach((id: ID) => {
      const element = subject.Elements.get(id);
      let res = undefined;

      switch (element.Type) {
        case RANK:
          const rank = element as RankQuestion;

          if (rankResult.has(id)) {
            res = countRankResult(rankResult.get(id), element as RankQuestion).resultsInPercent.map(
              (percent, index) => {
                return { Candidate: rank.Choices[index], Percentage: `${percent}%` };
              }
            );
            dataToDownload.push({ Title: element.Title.En, URL: element.Title.URL, Results: res });
          }
          break;

        case SELECT:
          const select = element as SelectQuestion;

          if (selectResult.has(id)) {
            res = countSelectResult(selectResult.get(id))
              .map(([, totalCount], index) => {
                return {
                  Candidate: select.Choices[index],
                  TotalCount: totalCount,
                  NumberOfBallots: selectResult.get(id).length, // number of combined ballots for this election
                };
              })
              .sort((x, y) => y.TotalCount - x.TotalCount);
            dataToDownload.push({ Title: element.Title.En, URL: element.Title.URL, Results: res });
          }
          break;

        case SUBJECT:
          getResultData(element as Subject, dataToDownload);
          break;

        case TEXT:
          if (textResult.has(id)) {
            res = Array.from(countTextResult(textResult.get(id)).resultsInPercent).map((r) => {
              return { Candidate: r[0], Percentage: `${r[1]}%` };
            });
            dataToDownload.push({ Title: element.Title.En, URL: element.Title.URL, Results: res });
          }
          break;
      }
    });
  };

  const exportJSONData = () => {
    const fileName = `result_${configuration.Title.En.replace(/[^a-zA-Z0-9]/g, '_').slice(
      0,
      99
    )}__grouped`; // replace spaces with underscores;

    const dataToDownload: DownloadedResults[] = [];

    configuration.Scaffold.forEach((subject: Subject) => {
      getResultData(subject, dataToDownload);
    });

    const data = {
      Title: configuration.Title,
      NumberOfVotes: result.length,
      Results: dataToDownload,
    };

    const fileToSave = new Blob([JSON.stringify(data, null, 2)], {
      type: 'application/json',
    });

    saveAs(fileToSave, fileName);
  };

  return !loading ? (
    <div>
      <div className="flex flex-col">
        {configuration.Scaffold.map((subject: Subject) => displayResults(subject))}
      </div>
      <div className="flex my-4"></div>
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

export default GroupedResult;
