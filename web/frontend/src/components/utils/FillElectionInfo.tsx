import { useEffect, useState } from 'react';
import { ID } from 'types/configuration';
import { ElectionInfo, LightElectionInfo, Result, STATUS } from 'types/electionInfo';

const useFillElectionInfo = (electionData: ElectionInfo) => {
  const [id, setId]: [ID, React.Dispatch<React.SetStateAction<ID>>] = useState(null);
  const [status, setStatus]: [STATUS, React.Dispatch<React.SetStateAction<STATUS>>] =
    useState(null);
  const [pubKey, setPubKey]: [string, React.Dispatch<React.SetStateAction<string>>] = useState('');
  const [result, setResult]: [Result[], React.Dispatch<React.SetStateAction<Result[]>>] =
    useState(null);
  const [chunksPerBallot, setChunksPerBallot]: [
    number,
    React.Dispatch<React.SetStateAction<number>>
  ] = useState(0);
  const [ballotSize, setBallotSize]: [number, React.Dispatch<React.SetStateAction<number>>] =
    useState(0);
  const [configObj, setConfigObj] = useState(null);
  const [isResultSet, setIsResultSet] = useState(false);

  useEffect(() => {
    if (electionData !== null) {
      setId(electionData.ElectionID);
      setStatus(electionData.Status);
      setPubKey(electionData.Pubkey);
      setResult(electionData.Result);
      setChunksPerBallot(electionData.ChunksPerBallot);
      setBallotSize(electionData.BallotSize);
      setConfigObj(electionData.Configuration);
      if (electionData.Result.length > 0) {
        setIsResultSet(true);
      }
    }
  }, [electionData]);

  return {
    id,
    status,
    setStatus,
    pubKey,
    result,
    setResult,
    chunksPerBallot,
    ballotSize,
    configObj,
    isResultSet,
    setIsResultSet,
  };
};

const useFillLightElectionInfo = (electionData: LightElectionInfo) => {
  const [id, setId] = useState(null);
  const [title, setTitle] = useState('');
  const [status, setStatus] = useState(null);
  const [pubKey, setPubKey] = useState('');

  useEffect(() => {
    if (electionData !== null) {
      setId(electionData.ElectionID);
      setTitle(electionData.Title);
      setStatus(electionData.Status);
      setPubKey(electionData.Pubkey);
    }
  }, [electionData]);

  return {
    id,
    title,
    status,
    setStatus,
    pubKey,
  };
};

export { useFillElectionInfo, useFillLightElectionInfo };
