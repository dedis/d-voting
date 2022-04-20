import React, { FC } from 'react';

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

type ResultProps = {
  resultData: Results[];
  configuration: Configuration;
};

// Functional component that display the result of the votes

const Result: FC<ResultProps> = ({ resultData, configuration }) => {
  //const [dataToDownload, setDataToDownload] = useState(null);

  const groupByID = (
    resultMap: Map<any, any>,
    IDs: ID[],
    results: any,
    toNumber: boolean = false
  ) => {
    IDs.map((id, index) => {
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
      // transform into an array of number to count them using reduce
      groupByID(selectResult, result.SelectResultIDs, result.SelectResult, true);
      groupByID(rankResult, result.RankResultIDs, result.RankResult);
      groupByID(textResult, result.TextResultIDs, result.TextResult);
    });

    console.log(selectResult);
    console.log(rankResult);
    console.log(textResult);

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
    /*if (dataToDownload === null) {
      setDataToDownload(sortedResultMap);
    }*/
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

  return (
    <div className="py-4">
      <h2 className="px-8 text-lg">Total number of votes : {resultData.length}</h2>
      {configuration.Scaffold.map((subject: Subject) => displayResults(subject))}
      {/*<DownloadResult resultData={dataToDownload}></DownloadResult>*/}
    </div>
  );
};

export default Result;
