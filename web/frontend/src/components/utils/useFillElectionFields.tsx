import { useState, useEffect } from 'react';

/*custom hook that given an election object returns its fields*/
const useFillElectionFields = (electionData) => {
  console.log('useFillElectionFields:', electionData);

  const [title, setTitle] = useState(null);
  const [candidates, setCandidates] = useState(null);
  const [id, setId] = useState(null);
  const [status, setStatus] = useState(null);
  const [pubKey, setPubKey] = useState('');
  const [result, setResult] = useState(null);
  const [isResultSet, setIsResultSet] = useState(false);

  useEffect(() => {
    if (electionData !== null) {
      setTitle(electionData.Title);
      setCandidates(electionData.Format.Candidates);
      setId(electionData.ElectionID);
      setStatus(electionData.Status);
      setPubKey(electionData.Pubkey);
      setResult(electionData.Result);
      if (electionData.Result.length > 0) {
        setIsResultSet(true);
      }
    }
  }, [electionData]);

  return {
    title,
    candidates,
    id,
    status,
    pubKey,
    result,
    setResult,
    setStatus,
    isResultSet,
    setIsResultSet,
  };
};

export default useFillElectionFields;
