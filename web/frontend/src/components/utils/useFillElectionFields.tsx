import { useEffect, useState } from 'react';
import { ElectionInfo } from 'types/frontendRequestBody';

const useFillElectionFields = (electionData: ElectionInfo) => {
  const [id, setId] = useState(null);
  const [status, setStatus] = useState(null);
  const [pubKey, setPubKey] = useState('');
  const [result, setResult] = useState(null);
  const [chunksPerBallot, setChunksPerBallot] = useState(0);
  const [ballotSize, setBallotSize] = useState(0);
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

export default useFillElectionFields;
