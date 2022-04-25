import React, { FC, useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { RankResults, SelectResults, TextResults } from 'types/electionInfo';
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
  TextQuestion,
} from 'types/configuration';
import DownloadButton from 'components/buttons/DownloadButton';
import { useTranslation } from 'react-i18next';
import saveAs from 'file-saver';
import { ROUTE_ELECTION_INDEX } from 'Routes';
import { Link, useParams } from 'react-router-dom';
import TextButton from '../../components/buttons/TextButton';
import useElection from 'components/utils/useElection';
import { useConfigurationOnly } from 'components/utils/useConfiguration';

// Functional component that display the result of the votes

const ElectionResult: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();

  const { loading, result, configObj } = useElection(electionId);
  const configuration = useConfigurationOnly(configObj);

  const [dataToDownload, setDataToDownload] = useState('');

  const [rankResult, setRankResult] = useState<RankResults>(null);
  const [selectResult, setSelectResult] = useState<SelectResults>(null);
  const [textResult, setTextResult] = useState<TextResults>(null);

  // Group the different results by the ID of the question SelectResult
  // are mapped to 0 or 1s, such that ballots can be count more efficiently
  const groupByID = (
    resultMap: Map<any, any>,
    IDs: ID[],
    results: any,
    toNumber: boolean = false
  ) => {
    IDs.forEach((id, index) => {
      let updatedRes: number[][] = new Array<number[]>();
      let res: number[] = results[index];

      if (toNumber) {
        res = results[index].map((r: boolean) => (r ? 1 : 0));
      }

      if (resultMap.has(id)) {
        updatedRes = resultMap.get(id);
      }

      updatedRes.push(res);
      resultMap.set(id, updatedRes);
    });
  };

  const groupResultsByID = () => {
    const selectRes: SelectResults = new Map<ID, number[][]>();
    const rankRes: RankResults = new Map<ID, number[][]>();
    const textRes: TextResults = new Map<ID, string[][]>();

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
        {element.Type === TEXT && (
          <TextResult text={element as TextQuestion} textResult={textResult.get(element.ID)} />
        )}
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

  const exportData = () => {
    const fileName = 'result.json';

    // Create a blob of the data
    const fileToSave = new Blob([JSON.stringify({ Result: result })], {
      type: 'application/json',
    });

    saveAs(fileToSave, fileName);
    //https://stackoverflow.com/questions/19721439/download-json-object-as-a-file-from-browser
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
            <Link to={ROUTE_ELECTION_INDEX}>
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
