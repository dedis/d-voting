import React, { FC, useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { DownloadedResults, RankResults, SelectResults, TextResults } from 'types/electionInfo';
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
import DownloadButton from 'components/buttons/DownloadButton';
import { useTranslation } from 'react-i18next';
import saveAs from 'file-saver';
import { Link, useParams } from 'react-router-dom';
import TextButton from '../../components/buttons/TextButton';
import useElection from 'components/utils/useElection';
import { useConfigurationOnly } from 'components/utils/useConfiguration';
import {
  countRankResult,
  countSelectResult,
  countTextResult,
} from './components/utils/countResult';

// Functional component that displays the result of the votes
const ElectionResult: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();

  const { loading, result, configObj } = useElection(electionId);
  const configuration = useConfigurationOnly(configObj);

  const [rankResult, setRankResult] = useState<RankResults>(null);
  const [selectResult, setSelectResult] = useState<SelectResults>(null);
  const [textResult, setTextResult] = useState<TextResults>(null);

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
      groupByID(selectRes, res.SelectResultIDs, res.SelectResult, true);
      groupByID(rankRes, res.RankResultIDs, res.RankResult);
      groupByID(textRes, res.TextResultIDs, res.TextResult);
    });

    return { rankRes, selectRes, textRes };
  };

  useEffect(() => {
    if (result !== null) {
      const { rankRes, selectRes, textRes } = groupResultsByID();

      setRankResult(rankRes);
      setSelectResult(selectRes);
      setTextResult(textRes);
    }
  }, [result]);

  const SubjectElementResultDisplay = (element: SubjectElement) => {
    return (
      <div className="px-4 pb-4">
        <h2 className="text-lg pb-2">{element.Title}</h2>
        {element.Type === RANK && (
          <RankResult rank={element as RankQuestion} rankResult={rankResult.get(element.ID)} />
        )}
        {element.Type === SELECT && (
          <SelectResult
            select={element as SelectQuestion}
            selectResult={selectResult.get(element.ID)}
          />
        )}
        {element.Type === TEXT && <TextResult textResult={textResult.get(element.ID)} />}
      </div>
    );
  };

  const displayResults = (subject: Subject) => {
    return (
      <div key={subject.ID} className="px-8 pt-4">
        <h2 className="text-lg font-bold">{subject.Title}</h2>
        {subject.Order.map((id: ID) => (
          <div key={id}>
            {subject.Elements.get(id).Type === SUBJECT ? (
              <div className="px-4">{displayResults(subject.Elements.get(id) as Subject)}</div>
            ) : (
              SubjectElementResultDisplay(subject.Elements.get(id))
            )}
          </div>
        ))}
      </div>
    );
  };

  const getResultData = (subject: Subject, dataToDownload: DownloadedResults[]) => {
    dataToDownload.push({ ID: subject.ID, Title: subject.Title });

    subject.Order.forEach((id: ID) => {
      const element = subject.Elements.get(id);
      let res = undefined;

      switch (element.Type) {
        case RANK:
          const rank = element as RankQuestion;
          res = countRankResult(rankResult.get(id), element as RankQuestion).resultsInPercent.map(
            (percent, index) => {
              return { Candidate: rank.Choices[index], Percentage: percent };
            }
          );
          dataToDownload.push({ ID: id, Title: element.Title, Results: res });
          break;

        case SELECT:
          const select = element as SelectQuestion;
          res = countSelectResult(selectResult.get(id)).resultsInPercent.map((percent, index) => {
            return { Candidate: select.Choices[index], Percentage: percent };
          });
          dataToDownload.push({ ID: id, Title: element.Title, Results: res });
          break;

        case SUBJECT:
          getResultData(element as Subject, dataToDownload);
          break;

        case TEXT:
          res = Array.from(countTextResult(textResult.get(id)).resultsInPercent).map((r) => {
            return { Candidate: r[0], Percentage: r[1] };
          });
          dataToDownload.push({ ID: id, Title: element.Title, Results: res });
          break;
      }
    });
  };

  const exportData = () => {
    const fileName = 'result.json';

    const dataToDownload: DownloadedResults[] = [];

    configuration.Scaffold.forEach((subject: Subject) => {
      getResultData(subject, dataToDownload);
    });

    const data = {
      Title: configuration.MainTitle,
      NumberOfVotes: result.length,
      Results: dataToDownload,
    };

    const fileToSave = new Blob([JSON.stringify(data, null, 2)], {
      type: 'application/json',
    });

    saveAs(fileToSave, fileName);
  };

  return (
    <div>
      {!loading ? (
        <div>
          <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            Results
          </h1>
          <div className="shadow-lg rounded-md w-full px-4 pb-4 my-0 sm:my-4">
            <h3 className="py-6 uppercase text-2xl text-center text-gray-700">
              {configuration.MainTitle}
            </h3>
            <h2 className="px-8 text-lg">Total number of votes : {result.length}</h2>
            {configuration.Scaffold.map((subject: Subject) => displayResults(subject))}
          </div>
          <div className="flex my-4">
            <Link to={'/elections/' + electionId}>
              <TextButton>{t('back')}</TextButton>
            </Link>
            <DownloadButton exportData={exportData}>{t('exportResJSON')}</DownloadButton>
          </div>
        </div>
      ) : (
        <p className="loading">{t('loading')}</p>
      )}
    </div>
  );
};

ElectionResult.propTypes = {
  location: PropTypes.any,
};

export default ElectionResult;
