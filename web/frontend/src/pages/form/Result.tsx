import { FC, useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { DownloadedResults, RankResults, SelectResults, TextResults } from 'types/form';
import { default as i18n } from 'i18next';
import {
  ID,
  RANK,
  RankQuestion,
  SELECT,
  SUBJECT,
  SelectQuestion,
  Subject,
  TEXT,
} from 'types/configuration';
import { useParams } from 'react-router-dom';
import { useNavigate } from 'react-router';
import useForm from 'components/utils/useForm';
import { useConfigurationOnly } from 'components/utils/useConfiguration';
import DownloadButton from 'components/buttons/DownloadButton';
import Loading from 'pages/Loading';
import saveAs from 'file-saver';
import ResultExplanation from './components/ResultExplanation';
import { Tab } from '@headlessui/react';
import GroupedResult from './GroupedResult';
import IndividualResult from './IndividualResult';
import {
  countRankResult,
  countSelectResult,
  countTextResult,
} from './components/utils/countResult';

// Functional component that displays the result of the votes
const FormResult: FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { formId } = useParams();

  const { loading, result, configObj } = useForm(formId);
  const configuration = useConfigurationOnly(configObj);

  const [rankResult, setRankResult] = useState<RankResults>(null);
  const [selectResult, setSelectResult] = useState<SelectResults>(null);
  const [textResult, setTextResult] = useState<TextResults>(null);

  const getResultData = (subject: Subject, dataToDownload: DownloadedResults[]) => {
    dataToDownload.push({ Title: subject.Title });

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
            dataToDownload.push({ Title: element.Title, Results: res });
          }
          break;

        case SELECT:
          const select = element as SelectQuestion;

          if (selectResult.has(id)) {
            res = countSelectResult(selectResult.get(id)).resultsInPercent.map((percent, index) => {
              return { Candidate: select.Choices[index], Percentage: `${percent}%` };
            });
            dataToDownload.push({ Title: element.Title, Results: res });
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
            dataToDownload.push({ Title: element.Title, Results: res });
          }
          break;
      }
    });
  };

  // Group the different results by the ID of the question,
  const groupByID = (
    resultMap: Map<ID, number[][] | string[][]>,
    IDs: ID[],
    results: boolean[][] | number[][] | string[][],
    toNumber: boolean = false
  ) => {
    IDs.forEach((id, index) => {
      let updatedRes = [];
      let res = results[index];

      // SelectResult are mapped to 0 or 1s, such that ballots can be
      // counted more efficiently
      if (toNumber) {
        res = (results[index] as boolean[]).map((r: boolean) => (r ? 1 : 0));
      }

      if (resultMap.has(id)) {
        updatedRes = resultMap.get(id);
      }

      updatedRes.push(res);
      resultMap.set(id, updatedRes);
    });
  };

  const groupResultsByID = () => {
    let selectRes: SelectResults = new Map<ID, number[][]>();
    let rankRes: RankResults = new Map<ID, number[][]>();
    let textRes: TextResults = new Map<ID, string[][]>();

    result.forEach((res) => {
      if (
        res.SelectResultIDs !== null &&
        res.RankResultIDs !== null &&
        res.TextResultIDs !== null
      ) {
        groupByID(selectRes, res.SelectResultIDs, res.SelectResult, true);
        groupByID(rankRes, res.RankResultIDs, res.RankResult);
        groupByID(textRes, res.TextResultIDs, res.TextResult);
      }
    });

    return { rankRes, selectRes, textRes };
  };
  const [titles, setTitles] = useState<any>({});
  useEffect(() => {
    try {
      console.log('config', configuration.MainTitle);
      const ts = JSON.parse(configuration.MainTitle);
      setTitles(ts);
    } catch (e) {
      console.log('error', e);
    }
  }, [configuration]);
  useEffect(() => {
    if (result !== null) {
      const { rankRes, selectRes, textRes } = groupResultsByID();

      setRankResult(rankRes);
      setSelectResult(selectRes);
      setTextResult(textRes);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [result]);

  const exportJSONData = () => {
    const fileName = 'result.json';

    const dataToDownload: DownloadedResults[] = [];

    configuration.Scaffold.forEach((subject: Subject) => {
      getResultData(subject, dataToDownload);
    });

    const data = {
      TitleEn: i18n.language == 'en' && titles.en,
      TitleFr: i18n.language == 'fr' && titles.fr,
      TitleDe: i18n.language == 'en' && titles.de,
      NumberOfVotes: result.length,
      Results: dataToDownload,
    };

    const fileToSave = new Blob([JSON.stringify(data, null, 2)], {
      type: 'application/json',
    });

    saveAs(fileToSave, fileName);
  };

  return (
    <div className="w-[60rem] font-sans px-4 pt-8 pb-4">
      {!loading ? (
        <div>
          <div className="flex items-center">
            <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
              {t('navBarResult')}
            </h1>
            <div className="pt-1.5">
              <ResultExplanation />
            </div>
          </div>
          <div className="w-full pb-4 my-0 sm:my-4">
            <h2 className="text-lg mt-2 sm:mt-4 sm:mb-6 mb-4">
              {t('totalNumberOfVotes', { votes: result.length })}
            </h2>
            <h3 className="py-6 border-t text-2xl text-center text-gray-700">
              {i18n.language === 'en' && titles.en}
              {i18n.language === 'fr' && titles.fr}
              {i18n.language === 'de' && titles.de}
            </h3>

            <div>
              <Tab.Group>
                <Tab.List className="flex space-x-1 rounded-xl p-1">
                  <Tab
                    key="grouped"
                    className={({ selected }) =>
                      selected
                        ? 'w-full rounded-lg py-2.5 text-sm font-medium leading-5 text-white bg-indigo-500 shadow'
                        : 'w-full rounded-lg py-2.5 text-sm font-medium leading-5 text-gray-700 hover:bg-indigo-100 hover:text-indigo-500'
                    }>
                    {t('resGroup')}
                  </Tab>
                  <Tab
                    key="individual"
                    className={({ selected }) =>
                      selected
                        ? 'w-full rounded-lg py-2.5 text-sm font-medium leading-5 text-white bg-indigo-500 shadow'
                        : 'w-full rounded-lg py-2.5 text-sm font-medium leading-5 text-gray-600 hover:bg-indigo-100 hover:text-indigo-500'
                    }>
                    {t('resIndiv')}
                  </Tab>
                </Tab.List>
                <Tab.Panel>
                  <GroupedResult
                    rankResult={rankResult}
                    selectResult={selectResult}
                    textResult={textResult}
                  />
                </Tab.Panel>
                <Tab.Panel>
                  <IndividualResult
                    rankResult={rankResult}
                    selectResult={selectResult}
                    textResult={textResult}
                    ballotNumber={result.length}
                  />
                </Tab.Panel>
              </Tab.Group>
            </div>
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
      )}
    </div>
  );
};

FormResult.propTypes = {
  location: PropTypes.any,
};

export default FormResult;
