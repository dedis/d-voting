import { useEffect, useState } from 'react';
import { ID } from 'types/configuration';
import { ElectionInfo, LightElectionInfo, Results, Status } from 'types/election';

const useFillElectionInfo = (electionData: ElectionInfo) => {
  const [id, setId] = useState<ID>('');
  const [status, setStatus] = useState<Status>(null);
  const [pubKey, setPubKey] = useState<string>('');
  const [roster, setRoster] = useState<string[]>(null);
  const [result, setResult] = useState<Results[]>(null);
  const [chunksPerBallot, setChunksPerBallot] = useState<number>(0);
  const [ballotSize, setBallotSize] = useState<number>(0);
  const [configObj, setConfigObj] = useState(null);
  const [isResultSet, setIsResultSet] = useState<boolean>(false);

  useEffect(() => {
    if (electionData !== null) {
      setId(electionData.ElectionID);
      setStatus(electionData.Status);
      setPubKey(electionData.Pubkey);
      setRoster(electionData.Roster);
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
    roster,
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
  const [id, setId] = useState<ID>('');
  const [title, setTitle] = useState<string>('');
  const [status, setStatus] = useState<Status>(null);
  const [pubKey, setPubKey] = useState<string>('');

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
