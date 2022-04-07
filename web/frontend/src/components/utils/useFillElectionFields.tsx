import { useEffect, useState } from 'react';

const useFillElectionFields = (electionData) => {
  const [electionTitle, setElectionTitle] = useState(null);
  const [id, setId] = useState(null);
  const [configObj, setConfigObj] = useState(null);
  const [status, setStatus] = useState(null);
  const [pubKey, setPubKey] = useState('');
  const [ballotSize, setBallotSize] = useState(0);
  const [chunksPerBallot, setChunksPerBallot] = useState(0);
  const [result, setResult] = useState(null);
  const [isResultSet, setIsResultSet] = useState(false);

  useEffect(() => {
    if (electionData !== null) {
      setElectionTitle(electionData.Title);
      setId(electionData.ElectionID);
      setConfigObj(electionData.Configuration);
      setStatus(electionData.Status);
      setPubKey(electionData.Pubkey);
      setBallotSize(electionData.BallotSize);
      setChunksPerBallot(electionData.ChunksPerBallot);
      setResult(electionData.Result);
      if (electionData.Result.length > 0) {
        setIsResultSet(true);
      }
    }
  }, [electionData]);

  return {
    electionTitle,
    id,
    configObj,
    status,
    pubKey,
    ballotSize,
    chunksPerBallot,
    result,
    setResult,
    setStatus,
    isResultSet,
    setIsResultSet,
  };
};

export default useFillElectionFields;
