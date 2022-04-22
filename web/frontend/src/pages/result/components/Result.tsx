import React, { FC, useState } from 'react';

//import DownloadResult from './DownloadResult';
import { RankResults, Results, SelectResults, TextResults } from 'types/electionInfo';
import SelectResult from './SelectResult';
import RankResult from './RankResult';
import TextResult from './TextResult';
import {
  Configuration,
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
import { Link } from 'react-router-dom';
import TextButton from '../../../components/buttons/TextButton';

type ResultProps = {
  resultData: Results[];
  configuration: Configuration;
};

// Functional component that display the result of the votes

const Result: FC<ResultProps> = ({ resultData, configuration }) => {
  const { t } = useTranslation();
  const [dataToDownload, setDataToDownload] = useState('');

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
    const selectResult: SelectResults = new Map<ID, number[][]>();
    const rankResult: RankResults = new Map<ID, number[][]>();
    const textResult: TextResults = new Map<ID, string[][]>();

    resultData.forEach((result) => {
      groupByID(selectResult, result.SelectResultIDs, result.SelectResult, true);
      groupByID(rankResult, result.RankResultIDs, result.RankResult);
      groupByID(textResult, result.TextResultIDs, result.TextResult);
    });

    return { rankResult, textResult, selectResult };
  };

  const { rankResult, textResult, selectResult } = groupResultsByID();

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
    const fileToSave = new Blob([JSON.stringify({ Result: resultData })], {
      type: 'application/json',
    });

    saveAs(fileToSave, fileName);
    //https://stackoverflow.com/questions/19721439/download-json-object-as-a-file-from-browser
  };

  return (
    <>
      <div className="shadow-lg rounded-md w-full px-4 pb-4 my-0 sm:my-4">
        <h3 className="py-6 uppercase text-2xl text-center text-gray-700">
          {configuration.MainTitle}
        </h3>
        <h2 className="px-8 text-lg">Total number of votes : {resultData.length}</h2>
        {configuration.Scaffold.map((subject: Subject) => displayResults(subject))}
        {/*TODO the ROUTE might need to be passed as a props */}
      </div>
      <div className="flex my-4">
        <Link to={ROUTE_ELECTION_INDEX}>
          <TextButton>{t('back')}</TextButton>
        </Link>
        <DownloadButton exportData={exportData}>{t('exportResJSON')}</DownloadButton>
      </div>
    </>
  );
};

export default Result;
